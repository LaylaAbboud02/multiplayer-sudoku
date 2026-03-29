console.log("Frontend JS loaded.");

const GAME_STATE_WAITING = "waiting";
const GAME_STATE_READY = "ready";
const GAME_STATE_IN_PROGRESS = "in_progress";
const GAME_STATE_FINISHED = "finished";

const MESSAGE_TYPE_ROOM_STATUS = "room_status";
const MESSAGE_TYPE_PLAYER_ASSIGNMENT = "player_assignment";

const MAX_MISTAKES = 4;
const ALLOWED_KEYS = [
  "Backspace",
  "Delete",
  "Tab",
  "ArrowLeft",
  "ArrowRight",
  "ArrowUp",
  "ArrowDown",
];

const inputs = document.querySelectorAll(".cell-input");
const mistakesDisplay = document.getElementById("mistakes-display");
const attemptsLeftDisplay = document.getElementById("attempts-left-display");
const statusMessage = document.getElementById("status-message");

const livePlayerCount = document.getElementById("live-player-count");
const liveRoomStatus = document.getElementById("live-room-status");
const playerRole = document.getElementById("player-role");
const gameStateDisplay = document.getElementById("game-state-display");

const boardWrapper = document.querySelector(".overflow-x-auto");
const pageRoot = document.querySelector("main");

let roomReady = pageRoot?.dataset.roomReady === "true";
let gameState = pageRoot?.dataset.gameState || GAME_STATE_WAITING;
let playerNumber = null;
let mistakes = 0;
let gameOver = false;

// Runs the first setup for the page:
// - updates the UI based on the initial values we got from HTML
// - attaches listeners to all Sudoku input cells
// - starts the WebSocket connection for live room updates
init();

function init() {
  updateMistakeUI();
  updateRoomReadyUI();
  updateGameStateUI();
  attachInputListeners();
  connectWebSocket();
}

function attachInputListeners() {
  inputs.forEach((input) => {
    input.addEventListener("keydown", handleCellKeydown);
    input.addEventListener("input", handleCellInput);
  });
}

// Small helper that tells us whether the player should be blocked from interacting.
// Interaction is locked if:
// - the game is over
// - or the room is not ready yet
function isInteractionLocked() {
  return gameOver || !roomReady;
}

function copyRoomCode() {
  const roomCode = document.getElementById("room-code")?.textContent?.trim();
  if (!roomCode) return;

  navigator.clipboard
    .writeText(roomCode)
    .then(() => {
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
    })
    .catch((err) => {
      console.error("Failed to copy:", err);
    });
}

// Handles keyboard presses before the character is actually inserted into the cell.
// We use this to block invalid keys early.
function handleCellKeydown(event) {
  // If the game already ended or the room is not ready, do not let the user type anything.
  if (isInteractionLocked()) {
    event.preventDefault();
    return;
  }

  // These keys are allowed because they help with editing/navigation.
  if (ALLOWED_KEYS.includes(event.key)) {
    return;
  }

  // For actual character input, only allow digits 1 through 9.
  // Anything else gets blocked before it appears in the input.
  if (!/^[1-9]$/.test(event.key)) {
    event.preventDefault();
  }
}

// Handles changes to a Sudoku cell's value.
// This is useful as a second safety layer, especially for paste or odd browser behavior.
function handleCellInput(event) {
  // If the game is over or the room is not ready, immediately clear anything typed.
  if (isInteractionLocked()) {
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
      `Wrong number. You have ${MAX_MISTAKES - mistakes} attempt(s) left before losing.`
    );

    // If the player reached the mistake limit, end the game.
    if (mistakes >= MAX_MISTAKES) {
      endGameLoss();
    }

    return;
  }

  // If the number was correct, keep it in the cell
  // and check whether the whole puzzle is now solved.
  checkForWin();
}

// Updates the live room data shown in the UI when a WebSocket room_status message arrives.
// This includes:
// - live player count
// - room ready state
// - game state
// - waiting/connected room message
function updateLiveRoomStatus(playerCount, newGameState = gameState) {
  if (livePlayerCount) {
    livePlayerCount.textContent = playerCount;
  }

  roomReady = playerCount >= 2;
  gameState = newGameState;

  updateRoomReadyUI();
  updateGameStateUI();

  if (liveRoomStatus) {
    if (playerCount < 2) {
      liveRoomStatus.textContent = "Waiting for another player to join...";
      liveRoomStatus.className = "mt-2 font-semibold text-amber-700";
    } else {
      liveRoomStatus.textContent = "Both players connected";
      liveRoomStatus.className = "mt-2 font-semibold text-emerald-700";
    }
  }
}

function updateRoomReadyUI() {
  const shouldDisableInputs = isInteractionLocked();

  inputs.forEach((input) => {
    input.disabled = shouldDisableInputs;
  });

  // Toggle blur effect on the board when the player should not be able to interact.
  // Right now that means:
  // - room not ready
  // - or game is over
  if (boardWrapper) {
    if (shouldDisableInputs) {
      boardWrapper.classList.add("blur-sm", "select-none", "pointer-events-none", "p-4");
    } else {
      boardWrapper.classList.remove("blur-sm", "select-none", "pointer-events-none", "p-4");
    }
  }

  // If the game is already over, leave the final status message alone.
  if (gameOver) return;

  if (!roomReady) {
    statusMessage.textContent = "Waiting for another player to join...";
  } else {
    statusMessage.textContent = "Game in progress";
  }
}

// Updates the visible match state text based on the current internal gameState value.
function updateGameStateUI() {
  if (!gameStateDisplay) return;

  switch (gameState) {
    case GAME_STATE_WAITING:
      gameStateDisplay.textContent = "Waiting";
      break;
    case GAME_STATE_READY:
      gameStateDisplay.textContent = "Ready";
      break;
    case GAME_STATE_IN_PROGRESS:
      gameStateDisplay.textContent = "In Progress";
      break;
    case GAME_STATE_FINISHED:
      gameStateDisplay.textContent = "Finished";
      break;
    default:
      gameStateDisplay.textContent = gameState;
  }
}

// Updates the player role text shown in the sidebar.
// If no assignment has arrived yet, it keeps showing "Assigning...".
function updatePlayerRoleUI() {
  if (!playerRole) return;

  if (playerNumber === null) {
    playerRole.textContent = "Assigning...";
  } else {
    playerRole.textContent = `Player ${playerNumber}`;
  }
}

function updateMistakeUI() {
  mistakesDisplay.textContent = `${mistakes} / ${MAX_MISTAKES}`;
  attemptsLeftDisplay.textContent = `${MAX_MISTAKES - mistakes}`;
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

      if (msg.type === MESSAGE_TYPE_ROOM_STATUS) {
        console.log("Updating live room status:", msg.player_count);
        updateLiveRoomStatus(msg.player_count, msg.game_state);
      }

      if (msg.type === MESSAGE_TYPE_PLAYER_ASSIGNMENT) {
        console.log("Received player assignment:", msg.player_number);
        playerNumber = msg.player_number;
        updatePlayerRoleUI();
      }
    } catch (error) {
      console.error("Failed to parse WebSocket message:", error);
    }
  });

  socket.addEventListener("close", (event) => {
    console.log("WebSocket closed.", event);

    roomReady = false;
    gameState = GAME_STATE_WAITING;

    updateRoomReadyUI();
    updateGameStateUI();

    if (playerRole) {
      playerRole.textContent = "Not Connected";
    }

    if (statusMessage) {
      statusMessage.textContent = "Connection lost or room is unavailable.";
    }
  });

  socket.addEventListener("error", (error) => {
    console.error("WebSocket error:", error);
  });
}