const newgamebtn = document.getElementById("new-game-btn");
if (newgamebtn) {
    newgamebtn.addEventListener("click", () => chooseMenu("new"));
}

const continuegamebtn = document.getElementById("continue-btn");
if (continuegamebtn) {
    continuegamebtn.addEventListener("click", () => chooseMenu("continue"));
}




// async function chooseMenu(gameType) { 
//     const userId = localStorage.getItem("userId");

//     // Debugging: Ensure userId and gameType are correct
//     console.log("Starting game with:");
//     console.log("User ID:", userId);
//     console.log("Game Type:", gameType);

//     if (!userId) {
//         alert("Please log in first.");
//         return;
//     }

//     try {
//         const payload = {
//             player_id: userId, 
//             type: gameType
//         };

//         console.log("Sending Payload:", JSON.stringify(payload)); // Debugging

//         const response = await fetch('http://localhost:8080/menu', {
//             method: 'POST',
//             headers: { 'Content-Type': 'application/json' },
//             body: JSON.stringify(payload),
//         });

//         const data = await response.json();
//         console.log("Server Response:", data); // Debugging

//         if (data.success) {
//             window.location.href = "index.html"; 
            
//         } else {
//             alert(data.message || 'Error starting the game.');
//         }
//     } catch (error) {
//         console.error('Error starting the game:', error);
//     }
// }
const PRIMARY_PORT = 8080;
const SECONDARY_PORT = 8081;

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


async function chooseMenu(gameType) { 
    const userId = localStorage.getItem("userId");

    // Debugging: Ensure userId and gameType are correct
    console.log("Starting game with:");
    console.log("User ID:", userId);
    console.log("Game Type:", gameType);

    if (!userId) {
        alert("Please log in first.");
        return;
    }

    try {
        const payload = {
            player_id: userId, 
            type: gameType
        };

        console.log("Sending Payload:", JSON.stringify(payload)); // Debugging

        const response = await fetchWithFallback("/menu", {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify(payload),
        });

        const data = await response.json();
        console.log("Server Response:", data); // Debugging

        if (data.success) {
            window.location.href = "index.html"; 
        } else {
            alert(data.message || 'Error starting the game.');
        }
    } catch (error) {
        console.error('Error starting the game:', error);
    }
}
