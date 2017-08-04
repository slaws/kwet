var loadingClass = 'mdl-progress mdl-js-progress mdl-progress__indeterminate'
$(window).load(function () {
  if (window.location.pathname.indexOf("/settings") != -1 && window.location.hash.indexOf('-panel') != -1) {
    $(".mdl-tabs__panel").removeClass('is-active')
    $(".mdl-tabs__tab").removeClass('is-active')
    $("a[href="+window.location.hash+"]").addClass('is-active')
    $(window.location.hash).addClass('is-active')
  }
  var dialog = document.querySelector('#delete-confirm');
  if (! dialog.showModal) {
    dialogPolyfill.registerDialog(dialog);
  }

  $(".delete-rule").on('click',function() {
    $("#deleterulename").text(this.getAttribute('data-rule'))
    document.querySelector("#deleterulename").setAttribute('data-rule',this.getAttribute('data-rule'))
    dialog.showModal();
  });

  dialog.querySelector('.close').addEventListener('click', function() {
    var rule = $("#deleterulename").text()
    $('.spin[data-rule="'+rule+'"]').addClass(loadingClass);
    componentHandler.upgradeElement(document.querySelector('.spin[data-rule="'+rule+'"]'))
    $.ajax({
      url: '/settings/hub/rule/'+rule,
      type: 'DELETE',
      success: function(result) {
        $('.spin[data-rule="'+rule+'"]').removeClass(loadingClass);
        componentHandler.upgradeElement(document.querySelector('.spin[data-rule="'+rule+'"]'))
        location.reload(true);
      },
      error: function(req, status,err) {
        $('.spin[data-rule="'+rule+'"]').removeClass(loadingClass);
        componentHandler.upgradeElement(document.querySelector('.spin[data-rule="'+rule+'"]'))
        var snackbarContainer = document.querySelector('#alert-snackbar');
        var data = {
          message: status + ' while deleting rule : ' + err,
          timeout: 10000,
          actionHandler: null,
          actionText: ''
        };
        snackbarContainer.MaterialSnackbar.showSnackbar(data);

      }
    });
    dialog.close();
  });
  dialog.querySelector('.cancel').addEventListener('click', function() {
    var rule = $("#deleterulename").text()
    $('.spin[data-rule="'+rule+'"]').removeClass(loadingClass);
    componentHandler.upgradeElement(document.querySelector('.spin[data-rule="'+rule+'"]'))
    dialog.close();
  });
});
