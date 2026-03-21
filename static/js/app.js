console.log("Frontend JS loaded.");

const inputs = document.querySelectorAll(".cell-input");
const mistakesDisplay = document.getElementById("mistakes-display");
const attemptsLeftDisplay = document.getElementById("attempts-left-display");
const statusMessage = document.getElementById("status-message");

let mistakes = 0;
const maxMistakes = 4;
let gameOver = false;

updateMistakeUI();

inputs.forEach((input) => {
  // Listen for keyboard presses before the character is actually inserted.
  // We use this to block invalid keys early.
  input.addEventListener("keydown", (event) => {
    // If the game already ended, do not let the user type anything else.
    if (gameOver) {
      event.preventDefault();
      return;
    }

    // These keys are allowed because they help with editing/navigation.
    const allowedKeys = [
      "Backspace",
      "Delete",
      "Tab",
      "ArrowLeft",
      "ArrowRight",
      "ArrowUp",
      "ArrowDown",
    ];

    // If the pressed key is one of the allowed control keys, let it through.
    if (allowedKeys.includes(event.key)) {
      return;
    }

    // For actual character input, only allow digits 1 through 9.
    // Anything else gets blocked before it appears in the input.
    if (!/^[1-9]$/.test(event.key)) {
      event.preventDefault();
    }
  });

  // Listen for changes to the input's value.
  // This is useful as a second safety layer, especially for paste or odd browser behavior.
  input.addEventListener("input", (event) => {
    // If the game is over, immediately clear anything typed.
    if (gameOver) {
      event.target.value = "";
      return;
    }

    // Remove any characters that are not digits 1-9.
    // Then keep only the first valid character, because a Sudoku cell should contain one number max.
    let value = event.target.value.replace(/[^1-9]/g, "");
    value = value.charAt(0) || "";
    event.target.value = value;

    // If the input is now empty, stop here.
    // This covers cases like deleting the value or entering something invalid.
    if (value === "") {
      return;
    }

    // Read the correct answer for this cell from the data-solution attribute in the HTML.
    const correctValue = Number(event.target.dataset.solution);

    // If the entered number is wrong:
    // - clear the cell
    // - add 1 mistake
    // - update the UI
    // - warn the player
    if (Number(value) !== correctValue) {
      event.target.value = "";
      mistakes += 1;
      updateMistakeUI();

      alert(
        `Wrong number. You have ${maxMistakes - mistakes} attempt(s) left before losing.`
      );

      // If the player reached the mistake limit, end the game.
      if (mistakes >= maxMistakes) {
        endGameLoss();
      }

      return;
    }

    // If the number was correct, keep it in the cell
    // and check whether the whole puzzle is now solved.
    checkForWin();
  });
});

function updateMistakeUI() {
  mistakesDisplay.textContent = `Mistakes: ${mistakes} / ${maxMistakes}`;
  attemptsLeftDisplay.textContent = `Attempts left: ${maxMistakes - mistakes}`;
}

function endGameLoss() {
  gameOver = true;
  statusMessage.textContent = "Game over";
  disableAllInputs();
  alert("Game over. You reached the maximum number of mistakes.");
}

function disableAllInputs() {
  inputs.forEach((input) => {
    input.disabled = true;
  });
}

function checkForWin() {
  for (const input of inputs) {
    const correctValue = Number(input.dataset.solution);

    if (Number(input.value) !== correctValue) {
      return;
    }
  }

  gameOver = true;
  statusMessage.textContent = "Puzzle solved!";
  disableAllInputs();
  alert("Congratulations! You solved the puzzle.");
}