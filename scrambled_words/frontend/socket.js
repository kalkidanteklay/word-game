const socket = new WebSocket('ws://localhost:8080/ws');
socket.onopen = () => console.log('WebSocket connection opened');
socket.onmessage = (event) => console.log('Message received:', event.data);
socket.onerror = (error) => console.error('WebSocket error:', error);
socket.onclose = () => console.log('WebSocket connection closed');