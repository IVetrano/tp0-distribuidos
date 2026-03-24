import socket
from .utils import Bet

AMOUNT_SIZE = 4

AGENCY_SIZE = 2
FIRST_NAME_SIZE = 24
LAST_NAME_SIZE = 20
DOCUMENT_SIZE = 12
BIRTH_DATE_SIZE = 4
NUMBER_SIZE = 4

BET_SIZE = AGENCY_SIZE + FIRST_NAME_SIZE + LAST_NAME_SIZE + DOCUMENT_SIZE + BIRTH_DATE_SIZE + NUMBER_SIZE

ACK_OK = 0
ACK_BAD_REQUEST = 1
ACK_INTERNAL_ERROR = 2

TYPE_SIZE = 1

TYPE_BET_BATCH = 0
TYPE_FINISH = 1
TYPE_QUERY = 2

WINNERS_NOT_READY = 0
WINNERS_READY = 1

class ProtocolError(Exception):
    pass

class Protocol:
    def __init__(self, client_socket):
        self._client_socket = client_socket
    
    def _receive_n_bytes(self, n):
        data = b''
        while len(data) < n:
            try:
                packet = self._client_socket.recv(n - len(data))
            except OSError as e:
                raise ProtocolError(f"Error receiving data: {e}")
            if not packet:
                raise ProtocolError("Connection closed by client")
            data += packet
        return data

    def _send_all(self, data):
        total_sent = 0
        while total_sent < len(data):
            try:
                sent = self._client_socket.send(data[total_sent:])
            except OSError as e:
                raise ProtocolError(f"Error sending data: {e}")
            if sent == 0:
                raise ProtocolError("Connection closed by client")
            total_sent += sent

    def deserialize_bet(self, data):
        offset = 0

        # Agency (2 bytes)
        agency = int.from_bytes(data[offset:offset+AGENCY_SIZE], byteorder='big')
        agency = str(agency)
        offset += AGENCY_SIZE

        # First Name (20 bytes)
        first_name = data[offset:offset+FIRST_NAME_SIZE].rstrip(b'\x00').decode('utf-8')
        offset += FIRST_NAME_SIZE

        # Last Name (20 bytes)
        last_name = data[offset:offset+LAST_NAME_SIZE].rstrip(b'\x00').decode('utf-8')
        offset += LAST_NAME_SIZE

        # Document (12 bytes)
        document = data[offset:offset+DOCUMENT_SIZE].rstrip(b'\x00').decode('utf-8')
        offset += DOCUMENT_SIZE

        # Birth Date (4 bytes)
        year = int.from_bytes(data[offset:offset+2], byteorder='big')
        month = int.from_bytes(data[offset+2:offset+3], byteorder='big')
        day = int.from_bytes(data[offset+3:offset+4], byteorder='big')
        birth_date = f"{year:04d}-{month:02d}-{day:02d}"
        offset += BIRTH_DATE_SIZE

        # Number (4 bytes)
        number = int.from_bytes(data[offset:offset+NUMBER_SIZE], byteorder='big')
        number = str(number)
        offset += NUMBER_SIZE

        return Bet(agency, first_name, last_name, document, birth_date, number)

    def receive_bet(self):
        data = self._receive_n_bytes(BET_SIZE)
        return self.deserialize_bet(data)
    
    def receive_amount(self):
        amount_data = self._receive_n_bytes(AMOUNT_SIZE)
        return int.from_bytes(amount_data, byteorder='big')

    def receive_n_bets(self, n):
        bets = []
        for _ in range(n):
            bet = self.receive_bet()
            bets.append(bet)
        return bets

    def send_ack(self, ack_code):
        self._send_all(bytes([ack_code]))

    def receive_type(self):
        type_data = self._receive_n_bytes(TYPE_SIZE)
        return type_data[0]

    def receive_agency_id(self):
        agency_id_data = self._receive_n_bytes(AGENCY_SIZE)
        return int.from_bytes(agency_id_data, byteorder='big')

    def send_winners_not_ready(self):
        self._send_all(bytes([WINNERS_NOT_READY]))

    def send_winners(self, winners):
        data = bytearray()

        # Winners ready
        data.append(WINNERS_READY)

        # Number of winners
        count = len(winners)
        data.extend(count.to_bytes(AMOUNT_SIZE, byteorder='big'))

        # Winners DNIs
        for winner in winners:
            data.extend(winner.encode('utf-8').ljust(DOCUMENT_SIZE, b'\x00'))
        
        self._send_all(bytes(data))

    def close(self):
        try:
            self._client_socket.close()
        except OSError as e:
            raise ProtocolError(f"Error closing connection: {e}")