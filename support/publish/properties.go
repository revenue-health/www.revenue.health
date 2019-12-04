package publish

import (
	"context"
	"fmt"
	"github.com/lectio/dropmark"
	"github.com/lectio/properties"
	"github.com/lectio/resource"
	"github.com/spf13/afero"
	"net/url"
	"path/filepath"
)

type dropmarkTagsProperty struct {
	PropName properties.PropertyName `json:"name"`
	Tags     []*dropmark.Tag         `json:"tags"`
	Slice    []string                `json:"slice"`
}

// NewDropmarkTagsProperty creates a new dropmark tags property
func NewDropmarkTagsProperty(name properties.PropertyName, tags []*dropmark.Tag) properties.Property {
	return &dropmarkTagsProperty{
		PropName: name,
		Tags:     tags,
		Slice:    asStringSlice(tags),
	}
}

func asStringSlice(tags []*dropmark.Tag) []string {
	if tags != nil {
		var slice []string
		for _, tag := range tags {
			slice = append(slice, tag.Name)
		}
		return slice
	}
	return nil
}

// Copy copies the key/value pair into the given map
func (p *dropmarkTagsProperty) Copy(ctx context.Context, m map[string]interface{}, options ...interface{}) {
	if p.Tags != nil {
		var tags []string
		for _, tag := range p.Tags {
			tags = append(tags, tag.Name)
		}
		m[string(p.PropName)] = tags
	}
}

// Name returns the property name
func (p *dropmarkTagsProperty) Name(context.Context) properties.PropertyName {
	return p.PropName
}

// AnyValue returns the property value useful when the type isn't important
func (p *dropmarkTagsProperty) AnyValue(context.Context) interface{} {
	return p.Slice
}

// Value returns the property value when the type is important
func (p *dropmarkTagsProperty) Value(context.Context) []string {
	return p.Slice
}

// downloadedResourceProperty implements ResourceProperty
type downloadedResourceProperty struct {
	name          properties.PropertyName
	downloadURL   *url.URL
	localHREF     string
	localFS       afero.Fs
	localFileName string
}

// NewDownloadedResourceProperty creates a new dropmark tags property
func NewDownloadedResourceProperty(name properties.PropertyName, featuredImageURL string, localHrefBase string, localFS afero.Fs, suggestFileName string) properties.Property {
	url, uerr := url.Parse(featuredImageURL)
	if uerr != nil {
		return &properties.DefaultTextProperty{name + "_error", fmt.Sprintf("URL parse error %q for featuredImageURL %q", uerr.Error(), featuredImageURL)}
	}

	extn := filepath.Ext(url.Path)
	localFileName := fmt.Sprintf("%s%s", suggestFileName, extn)
	localHREF := fmt.Sprintf("%s/%s", localHrefBase, localFileName)

	return &downloadedResourceProperty{
		name:          name,
		downloadURL:   url,
		localHREF:     localHREF,
		localFS:       localFS,
		localFileName: localFileName,
	}
}

// Copy copies the key/value pair into the given map
func (p *downloadedResourceProperty) Copy(ctx context.Context, m map[string]interface{}, options ...interface{}) {
	if writeCtx, ok := ctx.Value(inWriteContent).(writeContentContextValue); ok {
		// we pass in ourselves into PageFromURL because we implement resource.FileAttachmentCreator, which will
		// cause PageFromURL to use CreateFile and AutoAssignExtension
		_, err := writeCtx.publisher.ResourceFactory.PageFromURL(ctx, p.downloadURL.String(), p)
		if err != nil {
			m[string(p.name)+"_error_url"] = p.downloadURL.String()
			m[string(p.name)+"_error"] = err.Error()
			return
		}
		m[string(p.name)] = p.localHREF
	} else {
		m[string(p.name)] = p.localHREF
	}
}

// Name returns the property name
func (p *downloadedResourceProperty) Name(context.Context) properties.PropertyName {
	return p.name
}

// AnyValue returns the property value useful when the type isn't important
func (p *downloadedResourceProperty) AnyValue(context.Context) interface{} {
	return p.downloadURL
}

// CreateFile satisfies resource.FileAttachmentCreator
func (p *downloadedResourceProperty) CreateFile(context.Context, *url.URL, resource.Type) (afero.Fs, afero.File, error) {
	destFile, err := p.localFS.Create(p.localFileName)
	if err != nil {
		return p.localFS, nil, err
	}
	return p.localFS, destFile, nil
}

// AutoAssignExtension satisfies resource.FileAttachmentCreator
func (p *downloadedResourceProperty) AutoAssignExtension(context.Context, *url.URL, resource.Type) bool {
	return true
}
