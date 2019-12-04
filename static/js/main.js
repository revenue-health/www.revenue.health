(function () {
    var pathname = window.location.pathname;
    $(".navbar-li").removeClass("active");
    if (pathname == '/') {
        $("#nav_home").addClass("active");
    } else if (pathname.includes('/pdgm/')) {
        $("#nav_pdgm").addClass("active");
    } else if (pathname.includes('/blog/')) {
        $("#nav_blog").addClass("active");
    } else if (pathname.includes('/news/')) {
        $("#nav_news").addClass("active");
    } else if (pathname.includes('/faq/')) {
        $("#nav_faq").addClass("active");
    } else if (pathname.includes('/executive-team/')) {
        $("#nav_team").addClass("active");
    } else if (pathname.includes('/contact/')) {
        $("#nav_contact").addClass("active");
    } else {
        $("#nav_home").addClass("active");
    }
})(); 