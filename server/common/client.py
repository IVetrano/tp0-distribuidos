from . import protocol
from .utils import store_bets, load_bets, has_won
import logging

class Client:
    def __init__(self, client_socket):
        self._proto = protocol.Protocol(client_socket)
        self._last_msg_type = None

    def _get_agency_winners(self, agency_id):
        winners = []
        for bet in load_bets():
            if bet.agency == agency_id and has_won(bet):
                winners.append(bet.document)
        return winners

    def _handle_bet_batch(self):
        amount = self._proto.receive_amount()
        bets = self._proto.receive_n_bets(amount)
        store_bets(bets)
        logging.info(f"action: apuesta_recibida | result: success | cantidad: {amount}")

        self._proto.send_ack(protocol.ACK_OK)

    def _handle_finish(self, lottery_state):
        self._proto.send_ack(protocol.ACK_OK)
        lottery_state.agency_finished()

    def _handle_query(self, lottery_state):
        agency_id = self._proto.receive_agency_id()
    
        if not lottery_state.winners_ready():
            self._proto.send_winners_not_ready()
            return

        winners = self._get_agency_winners(agency_id)
        self._proto.send_winners(winners)

    def handle_connection(self, lottery_state):
        try:
            while True:
                try:
                    msg_type = self._proto.receive_type()
                    self._last_msg_type = msg_type
                except protocol.ProtocolError as e:
                    # EOF, client closed connection
                    break

                if msg_type == protocol.TYPE_BET_BATCH:
                    self._handle_bet_batch()
                
                elif msg_type == protocol.TYPE_FINISH:
                    self._handle_finish(lottery_state)
                    break
                
                elif msg_type == protocol.TYPE_QUERY:
                    self._handle_query(lottery_state)
                    break

        except ValueError as e:
            if self._last_msg_type in (protocol.TYPE_BET_BATCH, protocol.TYPE_FINISH):
                try:
                    self._proto.send_ack(protocol.ACK_BAD_REQUEST)
                except protocol.ProtocolError as e:
                    pass
            
        except protocol.ProtocolError as e:
            logging.error(f"action: receive_message | result: fail | type: protocol_error | error: {e}")

        except Exception as e:
            if self._last_msg_type in (protocol.TYPE_BET_BATCH, protocol.TYPE_FINISH):
                try:
                    self._proto.send_ack(protocol.ACK_INTERNAL_ERROR)
                except protocol.ProtocolError as e:
                    pass
        finally:
            self._proto.close()
    
    def close_connection(self):
        try:
            self._proto.close()
        except protocol.ProtocolError as e:
            logging.error(f"action: close_client_connection | result: fail | state: closing_client_sockets | error: {e}")
