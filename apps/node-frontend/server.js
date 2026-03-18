'use strict';

const { trace, SpanStatusCode } = require('@opentelemetry/api');
const express = require('express');

const app = express();
const PORT = process.env.PORT || 3000;
const BACKEND_URL = process.env.BACKEND_URL || 'http://go-backend:8080';

app.get('/roll', async (req, res) => {
  const span = trace.getActiveSpan();
  let response;

  try {
    response = await fetch(`${BACKEND_URL}/roll`);
  } catch (err) {
    span?.setStatus({ code: SpanStatusCode.ERROR, message: `${err.name}: ${err.message}` });
    return res.status(503).json({ error: 'backend unreachable' });
  }

  const body = await response.json().catch(() => ({}));

  if (!response.ok) {
    span?.setStatus({
      code: SpanStatusCode.ERROR,
      message: `HTTP ${response.status}: ${body.error ?? 'backend error'}`,
    });
    return res.status(502).json(body);
  }

  span?.setAttribute('dice.value', body.value);
  return res.json(body);
});

app.get('/health', (_req, res) => res.json({ status: 'ok' }));

app.listen(PORT, () => {
  console.log(`node-frontend listening on :${PORT}`);
});
