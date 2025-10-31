import express from 'express';
import cors from 'cors';
import http from 'http';
import { ExpressPeerServer } from 'peer';

const app = express();
const PORT = process.env.PORT || 9000;
const PATH = process.env.PEER_PATH || '/';

app.use(cors({ origin: true }));

// Create HTTP server and attach PeerServer correctly
const httpServer = http.createServer(app);
const peerServer = ExpressPeerServer(httpServer, {
  // Use root path here; we mount under PATH below
  path: PATH,
  proxied: false,
  debug: true,
  key: 'peerjs',
  allow_discovery: true,
});

peerServer.on('connection', (client) => {
  console.log('[peer] connection:', client.getId ? client.getId() : 'unknown');
});

peerServer.on('disconnect', (client) => {
  console.log('[peer] disconnect:', client.getId ? client.getId() : 'unknown');
});

peerServer.on('message', (client, message) => {
  console.log('[peer] message:', message?.type || message);
});

peerServer.on('error', (err) => {
  console.error('[peer] error:', err);
});

app.use(PATH, peerServer);

app.get('/', (_req, res) => {
  res.json({ ok: true, peer: { port: PORT, path: PATH } });
});

// Express error handler
// eslint-disable-next-line @typescript-eslint/no-unused-vars
app.use((err, _req, res, _next) => {
  console.error('[express] error:', err);
  try {
    res.status(500).json({ ok: false, error: String(err?.message || err) });
  } catch {}
});

httpServer.on('error', (err) => {
  console.error('[http] server error:', err);
});

httpServer.listen(PORT, () => {
  console.log(`PeerJS server running on http://localhost:${PORT}${PATH}`);
});


