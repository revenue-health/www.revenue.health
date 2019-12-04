package publish

import (
	"context"
	"fmt"
	"github.com/Machiel/slugify"
	"github.com/lectio/dropmark"
	"github.com/lectio/link"
	"github.com/lectio/markdown"
	"github.com/lectio/progress"
	"github.com/lectio/properties"
	"github.com/lectio/resource"
	"github.com/lectio/score"
	"github.com/lectio/text"
	"github.com/spf13/afero"
	"golang.org/x/xerrors"
	"sync"
)

type writeContentContextKey string
type writeContentContextValue struct {
	publisher *MarkdownPublisher
}

var inWriteContent = writeContentContextKey("InWriteContent")

// TestConfigurator is passed into options when we're testing the system
type TestConfigurator interface {
	StopAfterTestItemsCount(ctx context.Context) uint
	SimulateLinkScores(ctx context.Context) bool
}

// ContentLocator is passed into options if we want to override the content path
type ContentLocator interface {
	ContentPath(ctx context.Context) string
}

// ImagesLocator is passed into options if we want to override the images path and ref
type ImagesLocator interface {
	ImageDownloadPath(ctx context.Context) string
	ImageReferenceURL(ctx context.Context) string
}

// LinkScorer is passed into ptions if we want to score links
type LinkScorer interface {
	ScoreLink(context.Context, link.Link) (score.LinkScores, bool, error)
}

// MarkdownPublisher converts a Bookmarks source to Hugo content
type MarkdownPublisher struct {
	AsynchWorkers        uint
	ExceptionCollector   progress.ExceptionCollector
	LinkFactory          link.Factory
	ResourceFactory      resource.Factory
	StopAfterErrorsCount uint
	BoundedPR            progress.BoundedProgressReporter
	BasePathConfigurator markdown.BasePathConfigurator
	ContentFS            afero.Fs
	ImageCacheFS         afero.Fs
	ImageCacheRefURL     string
	Store                markdown.Store
	ContentFactory       markdown.ContentFactory
	PropertiesFactory    properties.Factory
	LinkScorer           LinkScorer
	TestConfigurator     TestConfigurator
}

// NewMarkdownPublisher returns a new Pipeline for this strategy
func NewMarkdownPublisher(ctx context.Context, asynchWorkers uint, linkFactory link.Factory, bpc markdown.BasePathConfigurator, options ...interface{}) (*MarkdownPublisher, error) {
	result := new(MarkdownPublisher)
	result.AsynchWorkers = asynchWorkers
	result.LinkFactory = linkFactory
	result.BasePathConfigurator = bpc
	result.BoundedPR = progress.NewSummaryReporter("")
	result.StopAfterErrorsCount = 10

	if err := result.initOptions(ctx, options...); err != nil {
		return result, err
	}

	return result, nil
}

func (p *MarkdownPublisher) initOptions(ctx context.Context, options ...interface{}) error {
	var err error

	for _, option := range options {
		if instance, ok := option.(ContentLocator); ok {
			if p.ContentFS, err = p.BasePathConfigurator.ComposePath(ctx, instance.ContentPath(ctx)); err != nil {
				return err
			}
		}
		if instance, ok := option.(ImagesLocator); ok {
			if p.ImageCacheFS, err = p.BasePathConfigurator.ComposePath(ctx, instance.ImageDownloadPath(ctx)); err != nil {
				return err
			}
			p.ImageCacheRefURL = instance.ImageReferenceURL(ctx)
		}
		if instance, ok := option.(progress.ExceptionCollector); ok {
			p.ExceptionCollector = instance
		}
		if instance, ok := option.(progress.BoundedProgressReporter); ok {
			p.BoundedPR = instance
		}
		if instance, ok := option.(markdown.ContentFactory); ok {
			p.ContentFactory = instance
		}
		if instance, ok := option.(markdown.Store); ok {
			p.Store = instance
		}
		if instance, ok := option.(properties.Factory); ok {
			p.PropertiesFactory = instance
		}
		if instance, ok := option.(resource.Factory); ok {
			p.ResourceFactory = instance
		}
		if instance, ok := option.(LinkScorer); ok {
			p.LinkScorer = instance
		}
		if instance, ok := option.(TestConfigurator); ok {
			p.TestConfigurator = instance
		}
	}

	const defaultContentPath = "content/news"
	const defaultImagePathPrefix = "img"
	if p.ContentFS == nil {
		if p.ContentFS, err = p.BasePathConfigurator.ComposePath(ctx, defaultContentPath); err != nil {
			return err
		}
	}
	if p.ImageCacheFS == nil {
		if p.ImageCacheFS, err = p.BasePathConfigurator.ComposePath(ctx, fmt.Sprintf("static/%s/%s", defaultImagePathPrefix, defaultContentPath)); err != nil {
			return err
		}
		p.ImageCacheRefURL = fmt.Sprintf("/%s/%s", defaultImagePathPrefix, defaultContentPath)
	}

	if p.Store == nil {
		p.Store = markdown.NewFileStore(p.ContentFactory, p.BasePathConfigurator)
	}

	if p.ContentFactory == nil {
		p.ContentFactory = markdown.TheContentFactory
	}

	if p.PropertiesFactory == nil {
		p.PropertiesFactory = p.ContentFactory.PropertiesFactory()
	}

	if p.ResourceFactory == nil {
		p.ResourceFactory = resource.NewFactory()
	}

	return nil
}

func (p *MarkdownPublisher) registerError(ctx context.Context, err error) bool {
	if p.ExceptionCollector != nil && err != nil {
		return p.ExceptionCollector.CollectError(ctx, err)
	}

	// no error rewriting or halting by default
	return true
}

func (p *MarkdownPublisher) maxErrorsReached(ctx context.Context) bool {
	if p.ExceptionCollector != nil {
		return p.ExceptionCollector.MaxErrorsCollected(ctx)
	}

	// no halting by default
	return false
}

func (p *MarkdownPublisher) registerWarning(ctx context.Context, code, message string) bool {
	if p.ExceptionCollector != nil {
		return p.ExceptionCollector.CollectWarning(ctx, code, message)
	}

	// no halting by default
	return true
}

// WriterPrimaryKey satisfies markdown.WriterIndexer
func (p *MarkdownPublisher) WriterPrimaryKey(ctx context.Context, content markdown.Content, options ...interface{}) string {
	ic := content.(markdown.IdentifiedContent)
	return ic.PrimaryKey()
}

// WriteToFileName satisfies markdown.WriterIndexer
func (p *MarkdownPublisher) WriteToFileName(ctx context.Context, content markdown.Content, options ...interface{}) (afero.Fs, string) {
	fileName := fmt.Sprintf("%s.md", p.WriterPrimaryKey(ctx, content))
	return p.ContentFS, fileName
}

func (p *MarkdownPublisher) publishItem(ctx context.Context, index uint, item *dropmark.Item) error {
	retain, traversedLink, err := p.LinkFactory.TraverseLink(ctx, item.Link)
	if err != nil {
		return err
	}
	if !retain {
		return nil
	}

	finalURL, uerr := traversedLink.FinalURL()
	if uerr != nil {
		return uerr
	}
	cleanTitle := text.TransformText(ctx, item.Name, func(context.Context, string) {}, text.RemovePipedSuffixFromText)
	slug := slugify.Slugify(link.GetSimplifiedHostnameWithoutTLD(finalURL) + "-" + cleanTitle)
	fm := p.PropertiesFactory.EmptyMutable(ctx)
	breadCrumbs := []*dropmark.Tag{}
    homeBreadCrumb := new(dropmark.Tag)
    homeBreadCrumb.Name = "Home"
    breadCrumbs = append(breadCrumbs, homeBreadCrumb)
    newsBreadCrumb := new(dropmark.Tag)
    newsBreadCrumb.Name = "News"
	breadCrumbs = append(breadCrumbs, newsBreadCrumb)
	titleBreadCrumb := new(dropmark.Tag)
    titleBreadCrumb.Name = cleanTitle
	breadCrumbs = append(breadCrumbs, titleBreadCrumb)
	
	breadCrumbLinks := []*dropmark.Tag{}
    homeBreadCrumbLink := new(dropmark.Tag)
    homeBreadCrumbLink.Name = "/"
    breadCrumbLinks = append(breadCrumbLinks, homeBreadCrumbLink)
    newsBreadCrumbLink := new(dropmark.Tag)
    newsBreadCrumbLink.Name = "/news"
	breadCrumbLinks = append(breadCrumbLinks, newsBreadCrumbLink)
	titleBreadCrumbLink := new(dropmark.Tag)
    titleBreadCrumbLink.Name = "/"
    breadCrumbLinks = append(breadCrumbLinks, titleBreadCrumbLink)

	fm.Add(ctx, "archetype", item.ContentArchetype)
	fm.Add(ctx, "title", cleanTitle)
	fm.Add(ctx, "description", item.Description)
	fm.Add(ctx, "slug", slug)
	fm.AddParsed(ctx, "date", item.UpdatedAt)
	fm.Add(ctx, "link", finalURL.String())
	fm.Add(ctx, "sourceBaseUrl", link.GetSimplifiedHostname(finalURL))
	fm.Add(ctx, "cleanUrl", item.Link)
	fm.Add(ctx, "layout", "news_single")
	fm.AddProperty(ctx, NewDropmarkTagsProperty("categories", item.Tags))
	fm.AddProperty(ctx, NewDownloadedResourceProperty("banner", item.ThumbnailURL, p.ImageCacheRefURL, p.ImageCacheFS, slug))
	fm.AddProperty(ctx, NewDropmarkTagsProperty("breadcrumbs", breadCrumbs))
	fm.AddProperty(ctx, NewDropmarkTagsProperty("breadcrumbLinks", breadCrumbLinks))
	content, ok, err := markdown.TheContentFactory.NewIdentifiedContent(ctx, slug, fm, []byte(item.Content))
	if err != nil {
		return xerrors.Errorf("Error creating content for item %d: %w", item.Index, err)
	}
	if !ok {
		return nil
	}

	// some properties, like downloadedResourceProperty and link scores, need additional context information so let's prepare to pass it in
	if item.Content != "" {
		writeCtx := context.WithValue(ctx, inWriteContent, writeContentContextValue{publisher: p})
		err = p.Store.WriteContent(writeCtx, p, content, nil)
		if err != nil {
			return xerrors.Errorf("Error writing markdown for item %d: %w", item.Index, err)
		}
	}
	return nil
}

// Publish takes a source apiEndpoint and import options to create markdown files
func (p *MarkdownPublisher) Publish(ctx context.Context, apiEndpoint string, options ...interface{}) error {
	if !dropmark.IsValidAPIEndpoint(apiEndpoint) {
		return fmt.Errorf("API endpoint %q is not a Dropmark URL", apiEndpoint)
	}
	dropmarkColl, err := dropmark.Import(ctx, apiEndpoint, options...)
	if err != nil {
		return err
	}

	c := dropmarkColl.(*dropmark.Collection)
	itemsCount := len(c.Items)
	p.BoundedPR.StartReportableActivity(ctx, fmt.Sprintf("Publishing %d Dropmark Links from %q", itemsCount, c.APIEndpoint), itemsCount)
	if p.AsynchWorkers > 0 {
		var wg sync.WaitGroup
		queue := make(chan int, p.AsynchWorkers)
		for index, item := range c.Items {
			if p.TestConfigurator != nil {
				if uint(index) > p.TestConfigurator.StopAfterTestItemsCount(ctx) {
					break
				}
			}
			if p.maxErrorsReached(ctx) {
				p.registerError(ctx, xerrors.Errorf("[item %d] maximum errors reached", index))
				break
			}

			wg.Add(1)
			go func(index int, item *dropmark.Item) {
				defer wg.Done()
				if piErr := p.publishItem(ctx, uint(index), item); piErr != nil {
					p.registerError(ctx, xerrors.Errorf("[item %d] error: %w", index, piErr))
				}
				queue <- index
			}(index, item)
		}
		go func() {
			defer close(queue)
			wg.Wait()
		}()
		for range queue {
			p.BoundedPR.IncrementReportableActivityProgress(ctx, 1)
		}
	} else {
		for index, item := range c.Items {
			if p.TestConfigurator != nil {
				if uint(index) > p.TestConfigurator.StopAfterTestItemsCount(ctx) {
					break
				}
			}

			if piErr := p.publishItem(ctx, uint(index), item); piErr != nil {
				if !p.registerError(ctx, xerrors.Errorf("[item %d] error: %w", index, piErr)) {
					break
				}
			}
			p.BoundedPR.IncrementReportableActivityProgress(ctx, 1)
		}
	}
	p.BoundedPR.CompleteReportableActivityProgress(ctx, fmt.Sprintf("Published %d Dropmark Links from %q", itemsCount, c.APIEndpoint))

	return nil
}
