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
    $('.submit-contact-form').prop('disabled', 'disabled');
    $('.submit-contact-form').addClass('isDisabled');
    $('#contact-form').on('blur keyup change', 'textarea,input', function (event) {
        validateForm('#contact-form');
    });

    function validateForm(id) {
        var valid = $(id).validate().checkForm();
        if (valid) {
            $('.submit-contact-form').prop('disabled', false);
            $('.submit-contact-form').removeClass('isDisabled');
        } else {
            $('.submit-contact-form').prop('disabled', 'disabled');
            $('.submit-contact-form').addClass('isDisabled');
        }
    }
})(); 