import { describe, it, expect } from 'vitest';
import type { Message, RoomCreatedData, TurnChangedData } from '../handlers/types';

// Simple tests that verify JSON message structure without mocking WebSocket
describe('WebSocket Message JSON Format Validation', () => {
  it('should validate room_created message JSON structure', () => {
    const message: Message = {
      type: 'room_created',
      data: {
        room_id: 'ABC123',
        peers: [
          {
            client_id: 'client1',
            display_name: 'TestUser',
            color: '#FF5733',
            total_turn_time: 0,
          },
        ],
        current_turn: null,
      } as RoomCreatedData,
    };

    // Verify message structure
    expect(message.type).toBe('room_created');
    expect(message.data.room_id).toBe('ABC123');
    expect(Array.isArray(message.data.peers)).toBe(true);
    expect(message.data.peers).toHaveLength(1);
    expect(message.data.peers[0]).toMatchObject({
      client_id: 'client1',
      display_name: 'TestUser',
      color: '#FF5733',
      total_turn_time: 0,
    });
    expect(message.data.current_turn).toBeNull();
  });

  it('should validate turn_changed message JSON structure with turn_start_time', () => {
    const message: Message = {
      type: 'turn_changed',
      data: {
        room_id: 'ABC123',
        current_turn: {
          client_id: 'client2',
          display_name: 'Player2',
          color: '#33FF57',
          total_turn_time: 0,
        },
        turn_start_time: 1234567890000,
      } as TurnChangedData,
    };

    expect(message.type).toBe('turn_changed');
    expect(message.data.room_id).toBe('ABC123');
    expect(message.data.current_turn).toMatchObject({
      client_id: 'client2',
      display_name: 'Player2',
      color: '#33FF57',
    });
    expect(message.data.turn_start_time).toBe(1234567890000);
    expect(typeof message.data.turn_start_time).toBe('number');
  });

  it('should validate turn_changed message with null current_turn', () => {
    const message: Message = {
      type: 'turn_changed',
      data: {
        room_id: 'ABC123',
        current_turn: null,
        turn_start_time: null,
      } as TurnChangedData,
    };

    expect(message.type).toBe('turn_changed');
    expect(message.data.current_turn).toBeNull();
    expect(message.data.turn_start_time).toBeNull();
  });

  it('should validate create_room request JSON structure', () => {
    const request = {
      type: 'create_room',
      data: {
        display_name: 'TestUser',
        color: '#FF5733',
      },
    };

    expect(request.type).toBe('create_room');
    expect(request.data.display_name).toBe('TestUser');
    expect(request.data.color).toBe('#FF5733');
    expect(request.data).not.toHaveProperty('room_id'); // Optional field
  });

  it('should validate start_turn request JSON structure', () => {
    const request = {
      type: 'start_turn',
      data: {
        current_turn: 'client1',
        new_turn: 'client2',
      },
    };

    expect(request.type).toBe('start_turn');
    expect(request.data.current_turn).toBe('client1');
    expect(request.data.new_turn).toBe('client2');
  });

  it('should validate end_turn request JSON structure (empty new_turn)', () => {
    const request = {
      type: 'start_turn',
      data: {
        current_turn: 'client1',
        new_turn: '',
      },
    };

    expect(request.type).toBe('start_turn');
    expect(request.data.current_turn).toBe('client1');
    expect(request.data.new_turn).toBe(''); // Empty means end turn
  });
});

