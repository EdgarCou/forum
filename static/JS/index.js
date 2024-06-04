function newPostPopUp(){
    document.querySelector('#buttonPost').style.display = 'none';
    document.querySelector('.posts').style.display = 'block';    
}

function newCommentPopUp(button){
    button.style.display = 'none';
    postId = button.getAttribute("data-post-id");
    document.querySelector('.comments'+postId).style.display = 'block';
}

function togglePopup(){
    document.querySelector('#buttonPost').style.display = 'block';
    document.querySelector('.posts').style.display = 'none';
}

function togglePopup2(button){
    button.style.display = 'block';
    postId = button.getAttribute("data-post-id");
    document.querySelector('.comments'+postId).style.display = 'none';
    document.querySelector('#commentButton'+postId).style.display = 'block';
}

function togglePopup3(){
    document.querySelector(".topics").style.display = 'none';
}

function newTopicPopup(){
    document.querySelector("#buttonTopic").style.display = 'none';
    document.querySelector(".topics").style.display = 'block';
}   

function displayComments(button){
    postId = button.getAttribute("data-post-id");
    document.querySelector('#comments'+postId).style.display = 'block';
    button.onclick = function(){
        document.querySelector('#comments'+postId).style.display = 'none';
        button.onclick = function(){
            displayComments(button);
        }
    }
}

var socket = new WebSocket("ws://localhost:8080/ws");

socket.onopen = function(event) {
    console.log("WebSocket is open now.");
    setupEventListeners();
};


socket.onmessage = function(event) {
    var data = event.data.split(":");
    console.log(data);
    if (data.length > 3) {
        var type1 = data[0];
        var postId1 = data[1];
        var count1 = data[2];
        var type2 = data[3];
        var postId2 = data[4];
        var count2 = data[5];
        if (type1 == 'likes') {
            document.getElementById("likeCount" + postId1).innerText = count1;
            document.getElementById("dislikeCount" + postId2).innerText = count2;
        } else if (type1 == 'dislikes') {
            document.getElementById("dislikeCount" + postId1).innerText = count1;
            document.getElementById("likeCount" + postId2).innerText = count2;
        }

    } else {
        var type = data[0];
        var postId = data[1];
        var count = data[2];
        if (type == 'likes') {
            document.getElementById("likeCount" + postId).innerText = count;
        } else if (type == 'dislikes') {
            document.getElementById("dislikeCount" + postId).innerText = count;
        }  
    }
    
};

function setupEventListeners() {
    var likeButtons = document.getElementsByClassName("likeButton");
    for (var i = 0; i < likeButtons.length; i++) {
        likeButtons[i].addEventListener("click", function(event) {
            event.preventDefault();
            var postId = this.getAttribute("data-post-id");
            socket.send("like:"+postId);
        });
    }

    var dislikeButtons = document.getElementsByClassName("dislikeButton");
    for (var i = 0; i < dislikeButtons.length; i++) {
        dislikeButtons[i].addEventListener("click", function(event) {
            event.preventDefault();
            var postId = this.getAttribute("data-post-id");
            socket.send("dislike:"+postId);
        });
    }
}

socket.onerror = function(event) {
    console.error("WebSocket error observed:", event);
};

socket.onclose = function(event) {
    console.log("WebSocket is closed now.", event);
};
