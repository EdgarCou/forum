function newPostPopUp(){
    document.querySelector('#buttonPost').style.display = 'none';
    document.querySelector('.posts').style.display = 'block';    
}

function togglePopup(){
    document.querySelector('#buttonPost').style.display = 'block';
    document.querySelector('.posts').style.display = 'none';
}

function likedButton(event){
    console.log("liked");
    var button = event.target;
    var otherButton = button.parentElement.querySelector('.dislikeButton');
    if (button.style.backgroundColor == 'red') {
        button.style.backgroundColor = 'white';
        otherButton.disabled = false;
    } else {
        button.style.backgroundColor = 'red';
        otherButton.disabled = true;
    }  
}

function dislikedButton(event){
    console.log("disliked");
    var button = event.target;
    var otherButton = button.parentElement.querySelector('.likeButton');
    if (button.style.backgroundColor == 'blue') {
        button.style.backgroundColor = 'white';
        otherButton.disabled = false;
    } else {
        button.style.backgroundColor = 'blue';
        otherButton.disabled = true;
    }  
}