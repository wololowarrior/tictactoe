const http = require('http');
const fs = require('fs');
const path = require('path');

const port = 3000;

const mimeTypes = {
    '.html': 'text/html',
    '.js': 'text/javascript',
    '.css': 'text/css',
    '.png': 'image/png',
    '.jpg': 'image/jpeg',
    '.gif': 'image/gif',
    '.ico': 'image/x-icon'
};

const server = http.createServer((req, res) => {
    console.log(`${new Date().toISOString()} - ${req.method} ${req.url}`);
    
    let filePath = req.url === '/' ? '/index.html' : req.url;
    filePath = path.join(__dirname, filePath);
    
    const extname = path.extname(filePath);
    const contentType = mimeTypes[extname] || 'text/plain';
    
    fs.readFile(filePath, (err, data) => {
        if (err) {
            if (err.code === 'ENOENT') {
                res.writeHead(404, { 'Content-Type': 'text/html' });
                res.end(`
                    <h1>404 Not Found</h1>
                    <p>The file ${req.url} was not found.</p>
                    <a href="/">Go back to home</a>
                `);
            } else {
                res.writeHead(500, { 'Content-Type': 'text/html' });
                res.end(`
                    <h1>500 Internal Server Error</h1>
                    <p>Server error: ${err.message}</p>
                `);
            }
        } else {
            res.writeHead(200, { 
                'Content-Type': contentType,
                'Access-Control-Allow-Origin': '*',
                'Access-Control-Allow-Methods': 'GET, POST, PUT, DELETE',
                'Access-Control-Allow-Headers': 'Content-Type'
            });
            res.end(data);
        }
    });
});

server.listen(port, () => {
    console.log('🌟 Nakama Tic-Tac-Toe Web UI Server');
    console.log('=====================================');
    console.log(`🚀 Server running at: http://localhost:${port}`);
    console.log(`📁 Serving files from: ${__dirname}`);
    console.log('🎮 Open the URL in your browser to play!');
    console.log('📝 Make sure Nakama server is running on localhost:7350');
    console.log('=====================================');
});

// Graceful shutdown
process.on('SIGTERM', () => {
    console.log('👋 Shutting down web server...');
    server.close(() => {
        console.log('✅ Server stopped');
        process.exit(0);
    });
});

process.on('SIGINT', () => {
    console.log('👋 Shutting down web server...');
    server.close(() => {
        console.log('✅ Server stopped');
        process.exit(0);
    });
});