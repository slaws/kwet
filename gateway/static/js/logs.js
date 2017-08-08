var scrolling=true;
var queue="";
var socket = new WebSocket("ws://"+location.host+"/wevents"+queue);

$(function () {
    var ws;
    if (window.WebSocket === undefined) {
        $("#container").append("Your browser does not support WebSockets");
        console.log("Your browser does not support WebSockets");
        return;
    } else {
        ws = initWS();
    }
    function initWS() {
        var container = $("#container")
        socket.onopen = function() {
          $("#socket-status").removeClass("socket-closed");
          $("#socket-status").addClass("socket-open");
            container.append("<p>Socket is open</p>");
        };
        socket.onmessage = function (e) {
          try {
            data = JSON.parse(e.data)
            // if(Array.isArray(data)) {
            //   // Syslog or fluentd errors
            //   $.each(data,function(i,data){
            //     var d = new Date(0);
            //     d.setUTCSeconds(data[0]);
            //     $('table tbody').append(`<tr>
            //       <td style=" word-wrap: break-word; white-space: normal; " class="mdl-data-table__cell--non-numeric time" >${d.toJSON()}</td>
            //       <td style=" word-wrap: break-word; white-space: normal; " class="mdl-data-table__cell--non-numeric">${data[1].tag}</td>
            //       <td style=" word-wrap: break-word; white-space: normal; " class="mdl-data-table__cell--non-numeric">${data[1].host}</td>
            //       <td style=" word-wrap: break-word; white-space: normal; " class="mdl-data-table__cell--non-numeric">${data[1].message}</td>
            //       </tr>
            //       `)
            //   });
            // } else {
            //   msg = JSON.parse(data.message)
            //   var d = new Date(0);
            //   d.setUTCSeconds(msg.lastseen);
            //   $('table tbody').append(`<tr>
            //     <td style=" word-wrap: break-word; white-space: normal; " class="mdl-data-table__cell--non-numeric time">${d.toJSON()}</td>
            //     <td style=" word-wrap: break-word; white-space: normal; " class="mdl-data-table__cell--non-numeric">${data.Subject}</td>
            //     <td style=" word-wrap: break-word; white-space: normal; " class="mdl-data-table__cell--non-numeric">${msg.source}</td>
            //     <td style=" word-wrap: break-word; white-space: normal; " class="mdl-data-table__cell--non-numeric">${msg.message}</td>
            //     </tr>
            //     `)
            // }
            var d = new Date();
            //d.setUTCSeconds(msg.lastseen);
            $('table tbody').append(`<tr>
              <td style=" word-wrap: break-word; white-space: normal; " class="mdl-data-table__cell--non-numeric time">${d.toJSON()}</td>
              <td style=" word-wrap: break-word; white-space: normal; " class="mdl-data-table__cell--non-numeric">${data.source}</td>
              <td style=" word-wrap: break-word; white-space: normal; " class="mdl-data-table__cell--non-numeric">${data.host}</td>
              <td style=" word-wrap: break-word; white-space: normal; " class="mdl-data-table__cell--non-numeric">${data.message}</td>
              </tr>
              `)

          }catch(e){
            console.log("Error ....")
            console.log(e)
          }
          if(scrolling) {
            $("main").scrollTop(document.getElementById("evtLogTable").offsetHeight);
          }
        }
        socket.onclose = function () {
          $("#socket-status").removeClass("socket-open");
          $("#socket-status").addClass("socket-closed");
            container.append("<p>Socket closed</p>");
          }
          return socket;
    }
});
function toggleScroll() {
  if(scrolling){
    $("#scroll-status").removeClass("scroll-on");
    $("#scroll-status").addClass("scroll-off");
    scrolling=false;
  }else{
    $("#scroll-status").removeClass("scroll-off");
    $("#scroll-status").addClass("scroll-on");
    scrolling=true;
  }
}
function watchQueue(q){
  //socket.send(q)
  q.length !== 0 ? $("#title").text(q) : $("#title").text("All queues")
  $('table tbody tr').slice(0).remove();
}
setInterval(function(){
  $.getJSON("/queues", function(list){
    var items = ['<a class="mdl-navigation__link" href="#" onclick="watchQueue(\'\')">All queues</a>'];
    $.each( list, function( key, val ) {
      items.push('<a class="mdl-navigation__link" href="#" onclick="watchQueue(\'' + val + '\')">'+ val+ '</a>');
    });
    $('#queuelist').html(items.join(""))
  })
}, 10000);

function toggleTime(){
    $('.time').toggle();
}

function fullscreen(){
  var elem = document.getElementById("main");
if (elem.requestFullscreen) {
  elem.requestFullscreen();
} else if (elem.mozRequestFullScreen) {
  elem.mozRequestFullScreen();
} else if (elem.webkitRequestFullscreen) {
  elem.webkitRequestFullscreen();
}
}

$(window).load(function () {
  var notification = document.querySelector('#snack .active');
  var data = {
    message: 'Connection to the bus messaging system failed.',
    actionHandler: function(event) {},
    actionText: 'Check Settings',
    timeout: 100000
  };
  if (notification != null) {
    notification.MaterialSnackbar.showSnackbar(data);
  }

  $("#search").on('keyup',function(e){
    if (e.which === 13){

      socket.send($('#search').val())
      $('table tbody tr').slice(0).remove();

    }
  })
});
