var scrolling=true;
var queue="";

$(function () {
    var ws;
    if (window.WebSocket === undefined) {
        $("#container").append("Your browser does not support WebSockets");
        return;
    } else {
        ws = initWS();
    }
    function initWS() {
        var socket = new WebSocket("ws://"+location.host+"/wevents"+queue),
            container = $("#container")
        socket.onopen = function() {
          $("#socket-status").removeClass("socket-closed");
          $("#socket-status").addClass("socket-open");
            container.append("<p>Socket is open</p>");
        };
        socket.onmessage = function (e) {
          try {
            data = JSON.parse(e.data)
            if(Array.isArray(data)) {
              // Syslog or fluentd errors
              console.log("array")
              $.each(data,function(i,data){
                var d = new Date(0);
                d.setUTCSeconds(data[0]);
                $('table tbody').append(`<tr>
                  <td style=" word-wrap: break-word; white-space: normal; " class="mdl-data-table__cell--non-numeric">${d.toJSON()}</td>
                  <td style=" word-wrap: break-word; white-space: normal; " class="mdl-data-table__cell--non-numeric">${data[1].tag}</td>
                  <td style=" word-wrap: break-word; white-space: normal; " class="mdl-data-table__cell--non-numeric">${data[1].host}</td>
                  <td style=" word-wrap: break-word; white-space: normal; " class="mdl-data-table__cell--non-numeric">${data[1].message}</td>
                  </tr>
                  `)
              });
            } else {
              console.log("ClusterEvent")
            }
            console.log(data)

          }catch(e){}
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
    console.log("Toggling off...");
    $("#scroll-status").removeClass("scroll-on");
    $("#scroll-status").addClass("scroll-off");
    scrolling=false;
  }else{
    console.log("Toggling on...");
    $("#scroll-status").removeClass("scroll-off");
    $("#scroll-status").addClass("scroll-on");
    scrolling=true;
  }
}
function watchQueue(q){
  console.log("Now watching queue "+ q);
}
