function newPostPopUp(){
    document.querySelector('.posts').style.display = 'block';   
    document.querySelector('#buttonTopic').disabled = true;
}

function togglePopup(){
    document.querySelector('.posts').style.display = 'none';
    document.querySelector('#buttonTopic').disabled = false;
}

function newCommentPopUp(button){
    postId = button.getAttribute("data-post-id");
    document.querySelector('.comments'+postId).style.display = 'block';
}

function togglePopup2(button){
    button.style.display = 'block';
    postId = button.getAttribute("data-post-id");
    document.querySelector('.comments'+postId).style.display = 'none';
    document.querySelector('#commentButton'+postId).style.display = 'block';
}

function newTopicPopUp(){
    document.querySelector('#wrapperTopic').style.display = 'block';
    document.querySelector('#buttonPost').disabled = true;
}   

function togglePopup3(){
    document.querySelector('#wrapperTopic').style.display = 'none';
    document.querySelector('#buttonPost').disabled = false;
}

function toggleInfo() {
    let popup = document.getElementById("info");
       popup.classList.toggle("open");
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

function myDropdownFunc(button) {
    document.getElementById("myDropdown").style.display = "block";
    button.onclick = function() {
        document.getElementById("myDropdown").style.display = "none";
        button.onclick = function() {
            myDropdownFunc(button);
        }
    }
}

function myDropdownFuncEdit(button) {
    document.getElementById("myDropdown-edit").style.display = "block";
    button.onclick = function() {
        document.getElementById("myDropdown-edit").style.display = "none";
        button.onclick = function() {
            myDropdownFuncEdit(button);
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

function displayTopicPosts(button){
    topicId = button.getAttribute("id");
    document.querySelector('#post_'+topicId).style.display = 'block';
    button.onclick = function(){
        document.querySelector('#post_'+topicId).style.display = 'none';
        button.onclick = function(){
            displayTopicPosts(button);
        }
    }
}

function deletePopUp(button){
    postId = button.getAttribute("data-post-id");
    console.log(postId);
    document.querySelector('#delete'+postId).style.display = 'block';
}

function editPopUp(button){
    postId = button.getAttribute("data-post-id");
    console.log(postId);
    document.querySelector('#edit'+postId).style.display = 'block';
    document.querySelector('#post'+postId).style.marginLeft = '60%';

}

function togglePopup4(button){
    postId = button.getAttribute("data-post-id");
    console.log(postId);
    document.querySelector('#delete'+postId).style.display = 'none';
}

function togglePopup5(button){
    postId = button.getAttribute("data-post-id");
    console.log(postId);
    document.querySelector('#edit'+postId).style.display = 'none';
    document.querySelector('.article-post').style.display = 'block';
}

function removeDelete(button){
    postId = button.getAttribute("data-post-id");
    document.querySelector('#delete'+postId).style.display = 'none';
}