console.log("Frontend JS loaded.");

const button = document.getElementById("test-btn");

if(button) {                
    button.addEventListener("click", () => {
        alert("Button was clicked!");
    }); 
}