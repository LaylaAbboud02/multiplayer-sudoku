console.log("Frontend JS loaded.");

const inputs = document.querySelectorAll(".cell-input");
const mistakesDisplay = document.getElementById("mistakes-display");
const attemptsLeftDisplay = document.getElementById("attempts-left-display");
const statusMessage = document.getElementById("status-message");

const livePlayerCount = document.getElementById("live-player-count");
const liveRoomStatus = document.getElementById("live-room-status");

const boardElement = document.querySelector(".board");
const boardWrapper = document.querySelector(".overflow-x-auto");

const pageRoot = document.querySelector("main");
let roomReady = pageRoot?.dataset.roomReady === "true";

let mistakes = 0;
const maxMistakes = 4;
let gameOver = false;

function copyRoomCode() {
  const roomCode = document.getElementById("room-code")?.textContent?.trim();
  if (!roomCode) return;

  navigator.clipboard.writeText(roomCode).then(() => {
    // Show copied state
    const copyIcon = document.querySelector(".copy-icon");
    const copiedIcon = document.querySelector(".copied-icon");
    if (copyIcon) copyIcon.classList.add("hidden");
    if (copiedIcon) copiedIcon.classList.remove("hidden");

    // Reset after 2 seconds
    setTimeout(() => {
      if (copyIcon) copyIcon.classList.remove("hidden");
      if (copiedIcon) copiedIcon.classList.add("hidden");
    }, 2000);
  }).catch((err) => {
    console.error("Failed to copy:", err);
  });
}

updateMistakeUI();
updateRoomReadyUI();
connectWebSocket();

inputs.forEach((input) => {
  // Listen for keyboard presses before the character is actually inserted.
  // We use this to block invalid keys early.
  input.addEventListener("keydown", (event) => {
    // If the game already ended, do not let the user type anything else.
    if (gameOver || !roomReady) {
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
    if (gameOver || !roomReady) {
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

function updateLiveRoomStatus(playerCount) {
  if(livePlayerCount) {
    livePlayerCount.textContent = playerCount;

  }

  roomReady = playerCount >= 2;
  updateRoomReadyUI();

  if(liveRoomStatus) {
    if(playerCount < 2){
      liveRoomStatus.textContent = "Waiting for another player to join...";
      liveRoomStatus.className = "mt-2 font-semibold text-amber-700";
    } else {
      liveRoomStatus.textContent = "Both players connected";
      liveRoomStatus.className = "mt-2 font-semibold text-emerald-700";
    }
  }
}

function updateRoomReadyUI() {
  const shouldDisableInputs = !roomReady || gameOver;

  inputs.forEach((input) => {
    input.disabled = shouldDisableInputs;
  });

  // Toggle blur effect on the board when not both players are connected
  if (boardWrapper) {
    if (!roomReady || gameOver) {
      boardWrapper.classList.add("blur-sm", "select-none", "pointer-events-none", "p-4");
    } else {
      boardWrapper.classList.remove("blur-sm", "select-none", "pointer-events-none", "p-4");
    }
  }

  if(gameOver) return;

  if(!roomReady){
    statusMessage.textContent = "Waiting for another player to join...";
  } else {
    statusMessage.textContent = "Game in progress";
  }
}

function updateMistakeUI() {
  mistakesDisplay.textContent = `Mistakes: ${mistakes} / ${maxMistakes}`;
  attemptsLeftDisplay.textContent = `Attempts left: ${maxMistakes - mistakes}`;
}

function endGameLoss() {
  gameOver = true;
  statusMessage.textContent = "Game over";
  updateRoomReadyUI();
  alert("Game over. You reached the maximum number of mistakes.");
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
  updateRoomReadyUI();
  alert("Congratulations! You solved the puzzle.");
}

function connectWebSocket() {
  if (typeof roomId === "undefined" || !roomId) {
    console.log("No roomId found, skipping WebSocket connection.");
    return;
  }

  const protocol = window.location.protocol === "https:" ? "wss:" : "ws:";
  const wsURL = `${protocol}//${window.location.host}/ws?room_id=${encodeURIComponent(roomId)}`;

  console.log("Attempting WebSocket connection...");
  console.log("roomId =", roomId);
  console.log("wsURL =", wsURL);

  const socket = new WebSocket(wsURL);

  socket.addEventListener("open", () => {
    console.log("WebSocket connected.");
  });

  socket.addEventListener("message", (event) => {
    console.log("WebSocket message received:", event.data);

    try {
      const msg = JSON.parse(event.data);

      if (msg.type === "room_status") {
        console.log("Updating live room status:", msg.player_count);
        updateLiveRoomStatus(msg.player_count);
      }
    } catch (error) {
      console.error("Failed to parse WebSocket message:", error);
    }
  });

  socket.addEventListener("close", (event) => {
    console.log("WebSocket closed.", event);
  });

  socket.addEventListener("error", (error) => {
    console.error("WebSocket error:", error);
  });
}