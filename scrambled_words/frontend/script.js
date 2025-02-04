let word = '';
let player_id = null; // Player's unique identifier
const PRIMARY_PORT = 8080;
const SECONDARY_PORT = 8081;
// Function to send a request with fallback
// Function to send a request with fallback
async function fetchWithFallback(url, options) {
    try {
        // Try the primary port first
        const response = await fetch(`http://localhost:${PRIMARY_PORT}${url}`, options);
        if (!response.ok) throw new Error("Primary port failed");
        return response;
    } catch (error) {
        console.error("Primary port failed, trying secondary port...");
        alert("Disconnected from the server. Reconnecting...");

        try {
            // If the primary port fails, try the secondary port
            const response = await fetch(`http://localhost:${SECONDARY_PORT}${url}`, options);
            if (!response.ok) throw new Error("Secondary port failed");

            alert("Reconnected to the server!");
            return response;
        } catch (error) {
            console.error("Secondary port failed.");
            alert("Unable to reconnect. Please try again later.");
            throw error; // Stop further retries
        }
    }
}
// Create input boxes for the current word
function createInputBoxes(word) {
    const inputContainer = document.getElementById("input_boxes");
    inputContainer.innerHTML = ''; 
    for (let char of word) {
        const input = document.createElement('input');
        input.type = 'text';
        input.maxLength = 1;
        input.className = 'input-box';
      

        input.addEventListener('input', (e) => {
            const nextInput = input.nextElementSibling;
            if (nextInput && nextInput.tagName === 'INPUT') {
                nextInput.focus();
            }
        });

        input.addEventListener('keydown', (e) => {
            if (e.key === 'Backspace' && input.value === '') {
                const previousInput = input.previousElementSibling;
                if (previousInput && previousInput.tagName === 'INPUT') {
                    previousInput.focus();
                }
            }
        });
        inputContainer.appendChild(input);
    }
}


window.onload = () => startGame(); 

async function startGame() {
    const userId = localStorage.getItem("userId");
    if (!userId) {
        alert("Please log in first.");
        return;
    }

    try {
        const payload = {
            player_id: userId, 
           
        };
        const response = await fetchWithFallback("/start", {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify(payload),
            
        });
        const data = await response.json();

        if (data.success) {
            player_id = data.player_id; // Assign the player ID
            word = data.word; // Get the word
            displayWord(word); // Show the word
        } else {
            alert(data.message || 'Error starting the game.');
        }
    } catch (error) {
        console.error('Error starting the game:', error);
    }
}




// Display the scrambled word
function displayWord(word) {
    const shuffledWord = shuffleString(word);
    const paragraph = document.getElementById("generated_text");

    // Ensure the element exists and is available
    if (paragraph) {
        const wrappedText = shuffledWord.split('').map(letter => {
            return `<span class="letters">${letter}</span>`;
        }).join('');

        // Safely set the innerHTML only when the element exists
        paragraph.innerHTML = wrappedText;

        // Proceed with creating input boxes for the word
        createInputBoxes(word);
    } else {
        console.error("The element with id 'generated_text' was not found.");
    }
}


    


// Shuffle a string (used for scrambling the word)
function shuffleString(str) {
    const arr = str.split(''); 
    for (let i = arr.length - 1; i > 0; i--) {
        const j = Math.floor(Math.random() * (i + 1)); 
        [arr[i], arr[j]] = [arr[j], arr[i]];
    }
    return arr.join(''); 
}

// Submit the player's guess
async function checkAnswer() {
    const userId = localStorage.getItem("userId");
    if (!userId) {
        alert("Please log in first.");
        return;
    }

    const inputBoxes = document.querySelectorAll('.input-box');
    let userGuess = '';

    inputBoxes.forEach(input => {
        userGuess += input.value.trim();
    });
    console.log("User ID:", userId); // Debugging
    console.log("User Guess:", userGuess); // Debugging

    try {
        const response = await fetchWithFallback("/submit",  {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({ player_id: userId, guess: userGuess }),
        });

        const data = await response.json();
        const playerScore = data.player ? data.player.score : "N/A"; 
        const resultMessage = document.getElementById("result_message");
        resultMessage.style.visibility = "visible";
        resultMessage.style.animation = "fadeIn 1s ease, zoom-in-zoom-out 1s ease infinite"; 
        setTimeout(() => {
            resultMessage.style.visibility = "hidden";
        }, 2000);

        if (data.correct) {
            resultMessage.textContent = "Correct!";
            resultMessage.style.color = "rgb(156, 236, 35);";
            const correctSound = document.getElementById("correct_sound");
                correctSound.play(); 
            word = data.new_word;
            
            
            displayWord(word); // Display the next word
        } else {
            
            resultMessage.textContent = "Incorrect!";
            resultMessage.style.color = "red";
            const IncorrectSound = document.getElementById("wrong-sound");
                IncorrectSound.play(); 
        }

        // Update the scores
        if (data.player) {
            updatePlayerList(data.player);
        }

        // Check for winner
        if (data.winner) {
            alert(`Game Over! ${data.winner} wins!`);
            window.location.reload(); // Reload the game
        }
    } catch (error) {
        console.error('Error checking answer:', error);
    }
}

// Update the score display
function updatePlayerList(players) {
    const playerListContainer = document.getElementById("player-list");
    
        playerListContainer.innerHTML = players
            .map(player => `${player.name}: ${player.score}`)
            .join("<br>");
        
    
}

// Event listeners
document.addEventListener("DOMContentLoaded", function () {
    const submitButton = document.getElementById("submit_button");
    if (submitButton) {
        submitButton.addEventListener("click", checkAnswer);
    }

    
    //window.onload = () => startGame(gameType); 
});


// Connect to WebSocket
// const socket = new WebSocket(`ws://localhost:${PRIMARY_PORT}/ws`);

// const name = localStorage.getItem("username");

// // Send username to server after connection is open
// socket.onopen = function () {
//     if (name) {
//         socket.send(JSON.stringify({
//             type: "register",
//             payload: {
//                 username: name,
//             },
//         }));
//     } else {
//         console.error("Username not found in local storage.");
//     }
// };

let socket = null;
let isPrimary = true; // Flag to indicate which port is currently being used

function connectWebSocket(port) {
    socket = new WebSocket(`ws://localhost:${port}/ws`);

    socket.onopen = function () {
        console.log(`Connected to WebSocket on port ${port}`);
        const name = localStorage.getItem("username");

        if (name) {
            socket.send(JSON.stringify({
                type: "register",
                payload: { username: name },
            }));
        } else {
            console.error("Username not found in local storage.");
        }

        
        const savedPlayers = localStorage.getItem("playerList");
        if (savedPlayers) {
            updatePlayerList(JSON.parse(savedPlayers));
        }
    };

    socket.onmessage = function (event) {
        const message = JSON.parse(event.data);
        console.log("Received message:", message);

        const gameover = document.getElementById("game-over-container");
        if (message.type === "player_list") {
            updatePlayerList(message.payload.players);
        }
        // Store updated player list in localStorage
        localStorage.setItem("playerList", JSON.stringify(message.payload.players));
        updatePlayerList(message.payload.players);

        if (message.type === "game_over") {
            const winner = message.payload.winner;
            const currentUser = localStorage.getItem("username");
            gameover.style.visibility = "visible";
            gameover.innerHTML = currentUser === winner ? "🎉🎉 YOU WON THE GAME!!" : "GAME OVER";
            gameover.style.color = "white";
            gameover.style.fontSize = "24px";
            gameover.style.textAlign = "center";
            gameover.style.fontWeight = "bold";

            if (currentUser !== winner) {
                alert(`${winner} won the game!`);
            }

            setTimeout(() => {
                window.location.href = "./menu.html";
            }, 4000);
        }
    };

    socket.onerror = function (error) {
        console.error(`WebSocket error on port ${port}:`, error);
    };

    socket.onclose = function () {
        console.warn(`WebSocket disconnected from port ${port}`);
        reconnectWebSocket();
    };
}

function reconnectWebSocket() {
    setTimeout(() => {
        if (isPrimary) {
            console.log("Trying to reconnect to secondary port...");
            isPrimary = false;
            connectWebSocket(SECONDARY_PORT);
        } else {
            console.log("Trying to reconnect to primary port...");
            isPrimary = true;
            connectWebSocket(PRIMARY_PORT);
        }
    }, 3000); // Retry after 3 seconds
}

// Start WebSocket connection on the primary port
connectWebSocket(PRIMARY_PORT);

// Handle WebSocket messages
socket.addEventListener('message', (event) => {
    const message = JSON.parse(event.data);

    console.log('Received message:', message);
    const gameover = document.getElementById("game-over-container");
   
    if (message.type === "player_list") {
        console.log("Updated player list:", message.payload.players);
        updatePlayerList(message.payload.players);  // Function to update UI
    }
    if (message.type === 'game_over') {
        const winner = message.payload.winner;
        const currentUser = localStorage.getItem("username");
        gameover.style.visibility = "visible";
        if (currentUser === winner) {
            gameover.innerHTML = "🎉🎉YOU WON THE GAME!!"
            gameover.style.color = "white";
          
            gameover.style.fontSize = "24px"; 
            gameover.style.textAlign = "center"; 
            gameover.style.fontWeight = "bold"; 
            //alert("You won the game!");
        } else {
            gameover.innerHTML = "GAME OVER"
            
            alert(`${winner} won the game!`);
            gameover.style.color = "white";
            gameover.style.margin = "20px";
            gameover.style.fontSize = "24px"; 
            gameover.style.textAlign = "center"; 
            gameover.style.fontWeight = "bold"; 
        }
        setTimeout(() => {
            window.location.href = "./menu.html";   
        }, 4000);
        // Redirect to the game-over screen
        
    }
});

// Handle WebSocket errors
socket.addEventListener('error', (error) => {
    console.error('WebSocket error:', error);
});

// Handle WebSocket close
socket.addEventListener('close', () => {
    console.log('WebSocket connection closed');
});