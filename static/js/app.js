console.log("Frontend JS loaded.");

const inputs = document.querySelectorAll(".cell-input");

inputs.forEach((input) => {
  input.addEventListener("input", (event) => {
    let value = event.target.value.trim();

    if (value.length > 1) {
      value = value.charAt(0);
    }

    if (!/^[1-9]?$/.test(value)) {
      value = "";
    }

    event.target.value = value;
  });
});