type MessageCallback = (message: any) => void;

interface Message {
  type: string;
  data: any;
}

// Connection Constants
const WS_URL = import.meta.env.VITE_WS_URL || "ws://localhost:8080/ws";
const RECONNECT_DELAY = 2000; // ms
const MAX_RECONNECT_ATTEMPTS = 5;
const CONNECTION_CHECK_INTERVAL = 100; // ms - how often to check connection state

// Message Waiting Constants
const DEFAULT_MESSAGE_TIMEOUT = 10000; // ms - used for all sendAndWait operations

export interface SendOptions {
  type: string;
  data: any;
}

export interface WaitOptions {
  type: string;
  validator?: (data: any) => boolean;
}

export class WebSocketConnection {
  private ws: WebSocket | null = null;
  private reconnectAttempts = 0;
  private reconnectTimer: NodeJS.Timeout | null = null;
  private messageCallbacks: Set<MessageCallback> = new Set();
  private isDestroyed = false;
  private connectPromise: Promise<void> | null = null;

  // Auto-connects if not already connected
  async connect(): Promise<void> {
    if (this.connectPromise) {
      return this.connectPromise;
    }

    if (this.ws?.readyState === WebSocket.OPEN) {
      return Promise.resolve();
    }

    this.connectPromise = new Promise((resolve, reject) => {
      if (this.ws?.readyState === WebSocket.CONNECTING) {
        const checkConnect = () => {
          if (this.ws?.readyState === WebSocket.OPEN) {
            this.connectPromise = null;
            resolve();
          } else if (this.ws?.readyState === WebSocket.CLOSED) {
            this.connectPromise = null;
            reject(new Error("Connection closed"));
          } else {
            setTimeout(checkConnect, CONNECTION_CHECK_INTERVAL);
          }
        };
        checkConnect();
        return;
      }

      try {
        this.ws = new WebSocket(WS_URL);

        this.ws.onopen = () => {
          console.log("WebSocket connected");
          this.reconnectAttempts = 0;
          this.connectPromise = null;
          resolve();
        };

        this.ws.onmessage = (event) => {
          try {
            const message: Message = JSON.parse(event.data);
            this.messageCallbacks.forEach((cb) => cb(message));
          } catch (err) {
            console.error("Error parsing message:", err);
          }
        };

        this.ws.onerror = (error) => {
          console.error("WebSocket error:", error);
          this.connectPromise = null;
          reject(error);
        };

        this.ws.onclose = () => {
          console.log("WebSocket closed");
          this.connectPromise = null;
          if (!this.isDestroyed) {
            this.attemptReconnect();
          }
        };
      } catch (err) {
        this.connectPromise = null;
        reject(err);
      }
    });

    return this.connectPromise;
  }

  // Auto-connects before sending
  async send(type: string, data: any): Promise<void> {
    await this.connect(); // Auto-connect if needed
    
    return new Promise((resolve, reject) => {
      if (!this.ws || this.ws.readyState !== WebSocket.OPEN) {
        reject(new Error("WebSocket not connected"));
        return;
      }

      const message: Message = { type, data };
      try {
        this.ws.send(JSON.stringify(message));
        resolve();
      } catch (err) {
        reject(err);
      }
    });
  }

  // Wait for a specific message type with optional validator
  // Uses DEFAULT_MESSAGE_TIMEOUT for all waits
  waitForMessage(
    messageType: string,
    validator?: (data: any) => boolean
  ): Promise<Message> {
    return new Promise((resolve, reject) => {
      // Set up one-time callback for this specific message
      const handler = (msg: Message) => {
        if (msg.type === messageType) {
          // If validator provided, check it
          if (validator && !validator(msg.data)) {
            return; // Keep waiting, this message doesn't match
          }
          
          // Found the message we're waiting for
          unsubscribe();
          clearTimeout(timeoutId);
          resolve(msg);
        } else if (msg.type === "error") {
          // Reject on any error message
          unsubscribe();
          clearTimeout(timeoutId);
          reject(new Error(msg.data?.message || "Server error"));
        }
      };

      const unsubscribe = this.onMessage(handler);

      // Set timeout using the default constant
      const timeoutId = setTimeout(() => {
        unsubscribe();
        reject(new Error(`${messageType} timeout after ${DEFAULT_MESSAGE_TIMEOUT}ms`));
      }, DEFAULT_MESSAGE_TIMEOUT);
    });
  }

  // Send a message and wait for a response - uses DEFAULT_MESSAGE_TIMEOUT
  async sendAndWait(
    send: SendOptions,
    wait: WaitOptions
  ): Promise<Message> {
    // Set up the wait promise BEFORE sending (to avoid race conditions)
    const waitPromise = this.waitForMessage(wait.type, wait.validator);
    
    // Auto-connect and send
    await this.send(send.type, send.data);
    
    // Wait for response (uses DEFAULT_MESSAGE_TIMEOUT internally)
    return waitPromise;
  }

  onMessage(callback: MessageCallback): () => void {
    this.messageCallbacks.add(callback);
    return () => {
      this.messageCallbacks.delete(callback);
    };
  }

  private attemptReconnect() {
    if (this.reconnectAttempts >= MAX_RECONNECT_ATTEMPTS) {
      console.error("Max reconnect attempts reached");
      return;
    }

    if (this.reconnectTimer) {
      clearTimeout(this.reconnectTimer);
    }

    this.reconnectAttempts++;
    console.log(`Attempting to reconnect (${this.reconnectAttempts}/${MAX_RECONNECT_ATTEMPTS})...`);

    this.reconnectTimer = setTimeout(() => {
      this.connect().catch((err) => {
        console.error("Reconnect failed:", err);
      });
    }, RECONNECT_DELAY);
  }

  destroy(): void {
    this.isDestroyed = true;

    if (this.reconnectTimer) {
      clearTimeout(this.reconnectTimer);
      this.reconnectTimer = null;
    }

    if (this.ws) {
      this.ws.close();
      this.ws = null;
    }

    this.messageCallbacks.clear();
    this.connectPromise = null;
  }
}

