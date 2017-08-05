var newlist = '<li class="mdl-list__item">              <div class="mdl-textfield mdl-js-textfield mdl-textfield--floating-label" style="width:40%; display: inline-block" >                <input class="mdl-textfield__input" type="text" id="newRuleAction" name="var[]" style="width:100%" >                <label class="mdl-textfield__label" for="newRuleAction">Key</label>              </div>              <div class="mdl-textfield mdl-js-textfield mdl-textfield--floating-label" style="width:40%;display:inline-block" >                <input class="mdl-textfield__input" type="text" id="newRuleAction" name="action[]"  style="width:100%" >                <label class="mdl-textfield__label" for="newRuleAction">Value</label>              </div>              <button href="#" type="button" class="delvar mdl-button mdl-js-button mdl-button--fab mdl-button--mini-fab">                <i class="material-icons">delete</i>              </button>            </li>';

$(window).load(function () {
  $("#newvar").on('click',function(e) {
    $("#varlist").append(newlist);
    componentHandler.upgradeDom();
  });
  $(document).on('click','.delvar',function(e) {
    console.log("true...true");
    $(this).closest('li').remove();
    componentHandler.upgradeDom();
  });
});
