from .protocol import Protocol, ProtocolError, ACK_OK, ACK_BAD_REQUEST, ACK_INTERNAL_ERROR
from .utils import store_bets
import logging

class Client:
    def __init__(self, client_socket):
        self._proto = Protocol(client_socket)

    def handle_connection(self):
        try:
            bet = self._proto.receive_bet()
            store_bets([bet])
            logging.info(
                "action: apuesta_almacenada | result: success | dni: %s | numero: %s",
                bet.document,
                bet.number,
            )

            self._proto.send_ack(ACK_OK)
        
        except ValueError as e:
            logging.error(f"action: receive_message | result: fail | type: bad_request | error: {e}")
            try:
                self._proto.send_ack(ACK_BAD_REQUEST)
            except ProtocolError as e:
                pass
        
        except ProtocolError as e:
            logging.error(f"action: receive_message | result: fail | type: protocol_error | error: {e}")

        except Exception as e:
            try:
                self._proto.send_ack(ACK_INTERNAL_ERROR)
            except ProtocolError as e:
                pass
        finally:
            self._proto.close()
    
    def close_connection(self):
        try:
            self._proto.close()
        except ProtocolError as e:
            logging.error(f"action: close_client_connection | result: fail | state: closing_client_sockets | error: {e}")
