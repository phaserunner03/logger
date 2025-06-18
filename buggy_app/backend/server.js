const express = require('express');
const fs = require('fs');
const cors = require('cors');
const app = express();
const PORT = 4000;

app.use(cors());
app.use(express.json());

app.post('/log-error', (req, res) => {
  const { message, stack } = req.body;

  const logEntry = {
    timestamp: new Date().toISOString(),
    text_payload: `${message}\n${stack}`
  };

  fs.appendFile('error_logs.json', JSON.stringify(logEntry) + '\n', err => {
    if (err) {
      console.error('Failed to write log:', err);
      res.status(500).send('Logging failed');
    } else {
      res.send('Logged');
    }
  });
});

app.listen(PORT, () => {
  console.log(`Error logging server running at http://localhost:${PORT}`);
});
