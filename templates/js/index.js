var listRooms = [];

window.onload = async function () {
  const userinfo = await loggedCheck();
  const token = userinfo['token'];
  const name = userinfo['name'];
  const username = userinfo['username'];
  const id = userinfo['id'];

  let conn;
  var msg = document.getElementById("msg");
  var log = document.getElementById("chat-box");
  const labelName = document.getElementById("label-name");
  const btnCreateRoom = document.getElementById("button-create-room");
  const chatContainer = document.getElementById("chat-container");
  const chatTitle = document.getElementById("chat-title");
  const btnLogout = document.getElementById("button-logout");

  btnCreateRoom.addEventListener('click', createRoom);
  btnLogout.addEventListener('click', logout);
  labelName.innerText = "Welcome, " + name;

  document.getElementById("form").onsubmit = function () {
    if (!conn) {
      return false;
    }
    if (!msg.value) {
      return false;
    }
    conn.send(`{"message":"${msg.value}","room_id":${msg.getAttribute('room_id')}}`);
    msg.value = "";
    return false;
  };


  //select2
  $('#select-usernames').select2({
    placeholder: "search user to start new chat",
    allowClear: true,
    minimumInputLength: 2,
    ajax: {
      url: window.location.origin + '/users',
      delay: 250,
      data: function (params) {
        var query = {
          s: params.term,
          type: 'select2'
        }
        return query;
      },
      headers: {
        "Authorization": token
      },
      processResults: function (data, params) {
        if (data['data'] == null) {
          return {
            results: []
          }
        } else {
          let result = data['data'].map((val, idx) => {
            return {
              id: val['username'],
              text: val['username'] + " - " + val['name']
            }
          });
          result = result.filter((val, idx) => {
            if (val.id != username) {
              return true;
            }
            return false;
          })
          return {
            results: result
          };
        }
      },
    }
  });


  async function loggedCheck() {
    const token = localStorage.getItem('token');
    if (token == null) {
      window.location.href = "/login";
    }

    const response = await fetch(window.location.origin + "/check-token", {
      method: 'GET',
      cache: 'no-cache',
      headers: {
        'Content-Type': 'application/json',
        'Authorization': token
      }
    });

    const payload = await response.json()
    return {
      token: token,
      id: payload.id,
      name: payload.name,
      username: payload.username
    }
  }

  function appendLog(item) {
    var doScroll = log.scrollTop > log.scrollHeight - log.clientHeight - 1;
    log.appendChild(item);
    if (doScroll) {
      log.scrollTop = log.scrollHeight - log.clientHeight;
    }
  }

  async function createRoom() {
    let participants = $("#select-usernames").val();
    const roomType = participants.length > 2 ? 'group' : 'private';
    if (participants.length > 0) {
      participants.push(username);
      let response = await fetch(window.location.origin + "/chats/room", {
        method: 'POST',
        cache: 'no-cache',
        headers: {
          'Content-Type': 'application/json',
          'Authorization': token
        },
        body: JSON.stringify({
          room_type: roomType,
          participants: participants
        })
      });
      response = await response.json();

      if (response["room_id"]) {
        listRooms.push({
          roomId: response["room_id"],
          roomType: roomType,
          participants: participants,
          createdBy: username,
          chats: []
        });
        chatContainer.style.display = "block";
        chatContainer.setAttribute("room_id", response["room_id"]);
        participants.splice(participants.indexOf(username), 1);
        chatTitle.innerText = participants.join(", ");
        msg.setAttribute('room_id', response["room_id"]);
        log.innerText = "";
      }
      $("#select-usernames").val([]).change();
    }
  }

  function connect() {
    if (window["WebSocket"]) {
      try {
        conn = new WebSocket("ws://" + document.location.host + "/ws?token=" + token);
      } catch (err) {
        console.log(err);
      }
      conn.onclose = function (evt) {
        setTimeout(function () {
          connect();
        }, 5000);
      };
      conn.onerror = function (err) {
        console.error('Socket encountered error: ', err.message, 'Closing socket');
        conn.close();
      };


      conn.onmessage = async function (evt) {
        var messages = evt.data.split('\n');
        for (var i = 0; i < messages.length; i++) {
          const parsed = JSON.parse(messages[i])
          const wrapper = document.createElement("div");
          const messageBox = document.createElement("div");
          const messageTitle = document.createElement("label");
          const messageContent = document.createElement("span");
          const roomId = parsed['Room'];

          let roomExist = false;
          listRooms.some(v => {
            if (v['roomId'] == roomId) {
              roomExist = true;
              v.chats.push(parsed)
              return true;
            }
            return false;
          });
          if (!roomExist) {
            const roomInfo = await getRoomInfo(roomId);
            listRooms.push({
              roomId: roomId,
              ...roomInfo,
              chats: [parsed]
            })
          }

          renderRooms();

          if (chatContainer.getAttribute("room_id") == roomId) {
            messageTitle.innerText = parsed['WriterName'];
            messageContent.innerText = parsed['Data'];

            messageBox.classList.add('mb-4', 'p-10', 'message-box', 'd-col');
            messageTitle.classList.add('message-title');
            messageContent.classList.add('message-content');
            if (parsed['WriterUsername'] == username) {
              wrapper.classList.add('wrapper-self')
              messageTitle.innerText = "You";
            } else {
              wrapper.classList.add('wrapper-other')
            }

            messageBox.appendChild(messageTitle);
            messageBox.appendChild(messageContent);
            wrapper.appendChild(messageBox);
            appendLog(wrapper);
          } else {
            const roomListItem = document.getElementById("room-list-item-" + roomId);
            const roomListItemBadge = document.createElement("span");
            roomListItemBadge.innerText = "new";
            roomListItem.appendChild(roomListItemBadge);
          }
        }
      };
    } else {
      var item = document.createElement("div");
      item.innerHTML = "<b>Your browser does not support WebSockets.</b>";
      appendLog(item);
    }
  }

  function renderRooms() {
    const divRoomList = document.getElementById("room-list");
    divRoomList.innerHTML = "";
    listRooms.forEach((v, i) => {
      const roomListItem = document.createElement("div");
      const roomListItemTitle = document.createElement("label");
      let participants = v["participants"];
      const myIndex = participants.indexOf(username);
      if (myIndex >= 0) {
        participants.splice(myIndex, 1);
      }
      roomListItem.classList.add("p-10", "mb-4", "d-row", "room-list-item");
      roomListItem.id = "room-list-item-" + v["roomId"];
      roomListItem.setAttribute('room_id', v["roomId"]);
      roomListItemTitle.innerText = participants.join(", ");
      roomListItem.appendChild(roomListItemTitle);
      roomListItem.addEventListener('click', showChatRoom)

      divRoomList.appendChild(roomListItem)
    })
  }

  function showChatRoom() {
    roomId = this.getAttribute("room_id");

    if (this.childElementCount > 1) {
      this.removeChild(this.childNodes[1]);
    }

    let roomIndex = -1;
    listRooms.some((val, idx) => {
      if (val["roomId"] == roomId) {
        roomIndex = idx;
        return true;
      }
      return false;
    })
    const data = listRooms[roomIndex];

    let participants = data["participants"];
    const myIndex = participants.indexOf(username);
    if (myIndex >= 0) {
      participants.splice(myIndex, 1);
    }
    chatContainer.style.display = "block";
    chatContainer.setAttribute("room_id", roomId);
    chatTitle.innerText = participants.join(", ");
    msg.setAttribute('room_id', data["roomId"]);
    log.innerText = "";
    data["chats"].forEach((v, i) => {
      const wrapper = document.createElement("div");
      const messageBox = document.createElement("div");
      const messageTitle = document.createElement("label");
      const messageContent = document.createElement("span");
      messageTitle.innerText = v['WriterName'];
      messageContent.innerText = v['Data'];

      messageBox.classList.add('mb-4', 'p-10', 'message-box', 'd-col');
      messageTitle.classList.add('message-title');
      messageContent.classList.add('message-content');
      if (v['WriterUsername'] == username) {
        wrapper.classList.add('wrapper-self')
        messageTitle.innerText = "You";
      } else {
        wrapper.classList.add('wrapper-other')
      }

      messageBox.appendChild(messageTitle);
      messageBox.appendChild(messageContent);
      wrapper.appendChild(messageBox);
      appendLog(wrapper);
    })

  }

  async function getRoomInfo(roomId) {
    const response = await fetch(window.location.origin + "/chats/room/" + roomId, {
      method: 'GET',
      cache: 'no-cache',
      headers: {
        'Content-Type': 'application/json',
        'Authorization': token
      },
    });
    const payload = await response.json()
    if (payload.status) {
      return payload.data;
    } else {
      return [];
    }
  }

  function logout(ev) {
    ev.preventDefault();
    localStorage.removeItem("token");
    window.location.href = "/login"
  }

  connect();
};
