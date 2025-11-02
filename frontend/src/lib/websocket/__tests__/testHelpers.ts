import type { Message } from '../handlers/types';

export class MockWebSocket {
  private listeners: Map<string, Set<Function>> = new Map();
  public readyState = WebSocket.OPEN;
  public messages: Message[] = [];
  public _listeners: Record<string, Function[]> = {};

  addEventListener(event: string, handler: Function) {
    if (!this._listeners[event]) {
      this._listeners[event] = [];
    }
    this._listeners[event].push(handler);
  }

  removeEventListener(event: string, handler: Function) {
    const handlers = this._listeners[event];
    if (handlers) {
      const index = handlers.indexOf(handler);
      if (index > -1) {
        handlers.splice(index, 1);
      }
    }
  }

  send(data: string) {
    try {
      const msg = JSON.parse(data);
      this.messages.push(msg);
    } catch {
      // Not JSON, ignore
    }
  }

  close() {
    this.readyState = WebSocket.CLOSED;
    this.trigger('close', {});
  }

  trigger(event: string, data: any) {
    const handlers = this._listeners[event] || [];
    handlers.forEach((handler) => {
      handler(data);
    });
  }

  simulateMessage(message: Message) {
    const event = {
      data: JSON.stringify(message),
    };
    this.trigger('message', event);
  }
}

export function createMockWebSocket(): MockWebSocket {
  return new MockWebSocket();
}

