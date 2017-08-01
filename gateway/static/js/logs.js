$(function () {
    var ws;
    if (window.WebSocket === undefined) {
        $("#container").append("Your browser does not support WebSockets");
        return;
    } else {
        ws = initWS();
    }
    function initWS() {
        var socket = new WebSocket("ws://"+location.host+"/wevents"),
            container = $("#container")
        socket.onopen = function() {
            container.append("<p>Socket is open</p>");
        };
        socket.onmessage = function (e) {
          try {
            data = JSON.parse(e.data)
            if(Array.isArray(data)) {
              // Syslog
              console.log("array")
              $.each(data,function(i,data){
                $('table tbody').append(`<tr>
                  <td class="mdl-data-table__cell--non-numeric">${data[0]}</td>
                  <td class="mdl-data-table__cell--non-numeric">${data[1].tag}</td>
                  <td class="mdl-data-table__cell--non-numeric">${data[1].message}</td>
                  </tr>
                  `)
              });
            } else {
              console.log("ClusterEvent")
            }
            console.log(data)
            msg = JSON.parse(data)
            console.log(msg)
            container.append("<p> Got data:" + msg.message + "</p>");
          }catch(e){}
        }
        socket.onclose = function () {
            container.append("<p>Socket closed</p>");
        }
        return socket;
    }
});
