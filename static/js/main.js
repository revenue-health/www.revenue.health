(function () {
    var pathname = window.location.pathname;
    //alert(pathname);
    $(".navbar-li").removeClass("active");
    if (pathname == '/') {
        $("#nav_wght_1").addClass("active");
    } else if (pathname.includes('/home-health/')) {
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
    } else if (pathname.includes('/contact/')) {
        $("#nav_wght_8").addClass("active");
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

$(document).ready(function(){
            $("a:contains('Read on')").each( function( i, element ) {
                $(this).attr('target','_blank');
             
         });
        // Add minus icon for collapse element which is open by default
        $(".collapse.show").each(function(){
            $(this).prev(".card-header").find(".fa").addClass("fa-minus").removeClass("fa-plus");
        });
        
        // Toggle plus minus icon on show hide of collapse element
        $(".collapse").on('show.bs.collapse', function(){
            $(this).prev(".card-header").find(".fa").removeClass("fa-plus").addClass("fa-minus");
        }).on('hide.bs.collapse', function(){
            $(this).prev(".card-header").find(".fa").removeClass("fa-minus").addClass("fa-plus");
        });
    });
