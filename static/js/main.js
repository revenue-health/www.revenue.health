(function () {
    var pathname = window.location.pathname;
    $(".navbar-li").removeClass("active");
    if (pathname == '/') {
        $("#nav_wght_1").addClass("active");
    } else if (pathname == '/home-health/') {
        $("#nav_wght_2").addClass("active");
    } else if (pathname.includes('/pdgm/')) {
        $("#nav_wght_3").addClass("active");
    } else if (pathname.includes('/blog/')) {
        $("#nav_wght_4").addClass("active");
    } else if (pathname.includes('/news/')) {
        $("#nav_wght_5").addClass("active");
    } else if (pathname.includes('/faq/')) {
        $("#nav_wght_6").addClass("active");
    } else if (pathname.includes('/executive-team/')) {
        $("#nav_wght_7").addClass("active");
    } else if (pathname.includes('/advisory-board/')) {
        $("#nav_wght_8").addClass("active");
    } else if (pathname.includes('/contact/')) {
        $("#nav_wght_9").addClass("active");
    } else {
        $("#nav_wght_1").addClass("active");
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