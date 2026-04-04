console.log("Frontend JS loaded.");

const GAME_STATE_WAITING = "waiting";
const GAME_STATE_READY = "ready";
const GAME_STATE_IN_PROGRESS = "in_progress";
const GAME_STATE_FINISHED = "finished";

const MESSAGE_TYPE_ROOM_STATUS = "room_status";
const MESSAGE_TYPE_PLAYER_ASSIGNMENT = "player_assignment";
const MESSAGE_TYPE_PLAYER_FINISHED = "player_finished";
const MESSAGE_TYPE_MATCH_RESULT = "match_result";
const MESSAGE_TYPE_PROGRESS_UPDATE = "progress_update";

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
const liveRoomSubtext = document.getElementById("live-room-subtext");

const livePlayerCount = document.getElementById("live-player-count");
const liveRoomStatus = document.getElementById("live-room-status");
const playerRole = document.getElementById("player-role");
const gameStateDisplay = document.getElementById("game-state-display");
const yourProgressDisplay = document.getElementById("your-progress-display");
const opponentProgressDisplay = document.getElementById("opponent-progress-display");

const boardWrapper = document.querySelector(".overflow-x-auto");
const pageRoot = document.querySelector("main");

let roomReady = pageRoot?.dataset.roomReady === "true";
let gameState = pageRoot?.dataset.gameState || GAME_STATE_WAITING;
let playerNumber = null;
let mistakes = 0;
let gameOver = false;

// This becomes true once the player solves the board and sends the finish message.
// We use it to temporarily lock the board while waiting for the server to announce the result.
let finishedSubmitted = false;

// We store the WebSocket here so other functions (like checkForWin) can send messages through it.
let socket = null;

// local race progress state
let myProgressCount = 0;
let opponentProgressCount = 0;

// Runs the first setup for the page:
// - updates the UI based on the initial values we got from HTML
// - attaches listeners to all Sudoku input cells
// - starts the WebSocket connection for live room updates
init();

function init() {
  updateMistakeUI();
  updateProgressUI();
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

  // If this cell was already solved before, do nothing.
  // This prevents counting the same cell twice.
  if (event.target.dataset.solved === "true") {
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

  // mark this cell as permanently solved so it cannot be counted again.
  event.target.dataset.solved = "true";
  event.target.disabled = true;

  // increase local progress and immediately update/broadcast it.
  myProgressCount += 1;
  updateProgressUI();
  sendProgressUpdate();

  // If the number was correct, keep it in the cell
  // and check whether the whole puzzle is now solved.
  checkForWin();
}

function sendPlayerFinished() {
  if (!socket || socket.readyState != WebSocket.OPEN) {
    console.error("WebSocket is not connected. Cannot send player finished message.");
    return;
  }

  const msg = {
    type: MESSAGE_TYPE_PLAYER_FINISHED
  };

  socket.send(JSON.stringify(msg));
  console.log("Sent player finished message to server.");
}

function sendProgressUpdate() {
  if (!socket || socket.readyState != WebSocket.OPEN) {
    console.error("WebSocket is not connected. Cannot send progress update.");
    return;
  }

  const msg = {
    type: MESSAGE_TYPE_PROGRESS_UPDATE,
    progress_count: myProgressCount
  }

  socket.send(JSON.stringify(msg));
  console.log("Sent progress update to server. Progress count:", myProgressCount);
}

function updateProgressUI() {
  const totalFillableCells = inputs.length;

  if (yourProgressDisplay) {
    yourProgressDisplay.textContent = `Your progress: ${myProgressCount} / ${totalFillableCells}`;
  }

  if (opponentProgressDisplay) {
    opponentProgressDisplay.textContent = `Opponent's progress: ${opponentProgressCount} / ${totalFillableCells}`;
  }
}

function handleProgressUpdate(player1Progress, player2Progress) {
  if (playerNumber === 1) { 
    myProgressCount = player1Progress;
    opponentProgressCount = player2Progress;
  } else if (playerNumber === 2) {
    myProgressCount = player2Progress;
    opponentProgressCount = player1Progress;
  } else {
    myProgressCount = 0;
    opponentProgressCount = 0;
  }

  updateProgressUI();
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
    const isDark = document.documentElement.classList.contains("dark");
    const isWinx = document.documentElement.classList.contains("winx");

    // ADDED: finished state should not look like a normal waiting lobby
    if (gameState === GAME_STATE_FINISHED) {
      liveRoomStatus.textContent = "Match finished.";

      if (isWinx) {
        liveRoomStatus.className = "font-semibold text-pink-600";
      } else if (isDark) {
        liveRoomStatus.className = "font-semibold text-zinc-300";
      } else {
        liveRoomStatus.className = "font-semibold text-zinc-700";
      }

    } else if (playerCount < 2) {
      liveRoomStatus.textContent = "Waiting for another player to join...";

      if (isWinx) {
        liveRoomStatus.className = "font-semibold text-pink-600";
      } else if (isDark) {
        liveRoomStatus.className = "font-semibold text-amber-400";
      } else {
        liveRoomStatus.className = "font-semibold text-amber-700";
      }

    } else {
      liveRoomStatus.textContent = "Both players connected.";

      if (isWinx) {
        liveRoomStatus.className = "font-semibold text-pink-600";
      } else if (isDark) {
        liveRoomStatus.className = "font-semibold text-emerald-400";
      } else {
        liveRoomStatus.className = "font-semibold text-emerald-700";
      }
    }
  }

    // ADDED: update the smaller room status text under the main status line
  if (liveRoomSubtext) {
    if (gameState === GAME_STATE_FINISHED) {
      liveRoomSubtext.textContent = "This match is over. Start a new room to play again.";
      console.log("Game finished, updated live room subtext accordingly.");
    } else if (playerCount < 2) {
      liveRoomSubtext.textContent = "Your room is live. Share the code and wait for someone to join.";
      console.log("playercount < 2, updated live room subtext accordingly.");
    } else {
      liveRoomSubtext.textContent = "Nice. The room is full and ready for the match.";
      console.log("playercount >= 2, updated live room subtext accordingly.");
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

  if(finishedSubmitted) {
    statusMessage.textContent = "Board solved! Waiting for result...";
    return;
  }

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
   if(finishedSubmitted) {
    return;
  
  }
  
  if (myProgressCount !== inputs.length) {
    return;
  }

  // At this point, the player solved the whole board.
  // We do NOT decide the winner locally anymore.
  // Instead, we tell the server and wait for the official result.
  finishedSubmitted = true;
  updateRoomReadyUI();
  sendPlayerFinished();
}

function handleMatchResult(winnerPlayerNumber) {
  gameOver = true;
  finishedSubmitted = false;
  gameState = GAME_STATE_FINISHED;

  updateGameStateUI();
  updateRoomReadyUI();

  if (winnerPlayerNumber === playerNumber) {
    statusMessage.textContent = "You won!";
    alert("You won! You finished the board first.");
  } else {
    statusMessage.textContent = "You lost.";
    alert("You lost. Your opponent finished first.");
  }
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

  socket = new WebSocket(wsURL);

  socket.addEventListener("open", () => {
    console.log("WebSocket connected.");
    sendProgressUpdate();
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

      if (msg.type === MESSAGE_TYPE_PROGRESS_UPDATE) {
        console.log("Received progress update. Player 1 progress:", msg.player1_progress, "Player 2 progress:", msg.player2_progress);
        handleProgressUpdate(msg.player1_progress, msg.player2_progress);
      }

      if(msg.type === MESSAGE_TYPE_MATCH_RESULT) {
        console.log("Received match result. Winner player number:", msg.winner_player_number);
        handleMatchResult(msg.winner_player_number);
      }
    } catch (error) {
      console.error("Failed to parse WebSocket message:", error);
    }
  });

  socket.addEventListener("close", (event) => {
    console.log("WebSocket closed.", event);

    roomReady = false;
    finishedSubmitted = false;
    
    if (!gameOver) {
      gameState = GAME_STATE_WAITING;
    }

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