function newPostPopUp(){
    document.querySelector('#buttonPost').style.display = 'none';
    document.querySelector('.posts').style.display = 'block';    
}

function togglePopup(){
    document.querySelector('#buttonPost').style.display = 'block';
    document.querySelector('.posts').style.display = 'none';
}