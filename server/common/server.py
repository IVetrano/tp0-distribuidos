import socket
import logging
from .client import Client


class LotteryState:
    def __init__(self, expected_agencies):
        self._expected_agencies = expected_agencies
        self._agencies_finish = 0
    
    def agency_finished(self):
        self._agencies_finish += 1
        if self._agencies_finish == self._expected_agencies:
            logging.info("action: sorteo | result: success")

    def winners_ready(self):
        return self._agencies_finish >= self._expected_agencies

class Server:
    def __init__(self, port, listen_backlog, expected_agencies):
        # Initialize server socket
        self._server_socket = socket.socket(socket.AF_INET, socket.SOCK_STREAM)
        self._server_socket.bind(('', port))
        self._server_socket.listen(listen_backlog)
        self._running = True
        self._clients = set()
        self._lottery_state = LotteryState(expected_agencies)

    def shutdown(self):
        self._running = False

        logging.info("action: shutdown | result: in_progress")
        if self._server_socket:
            try:
                self._server_socket.close()
                logging.info("action: shutdown | result: success | state: closing_server_socket")
            except OSError as e:
                logging.error(f"action: shutdown | result: fail | state: closing_server_socket | error: {e}")

        for client in self._clients.copy():
            client.close_connection()
        
        logging.info("action: shutdown | result: success")

    def run(self):
        """
        Dummy Server loop

        Server that accept a new connections and establishes a
        communication with a client. After client with communucation
        finishes, servers starts to accept new connections again
        """
        while self._running:
            try:
                client_sock = self.__accept_new_connection()
                if client_sock:
                    client = Client(client_sock)
                    self._clients.add(client)
                    self.__handle_client_connection(client)
            except OSError as e:
                if not self._running:
                    break
                logging.error(f"action: accept_connections | error: {e}")

    def __handle_client_connection(self, client):
        """
        Read message from a specific client socket and closes the socket

        If a problem arises in the communication with the client, the
        client socket will also be closed
        """
        client.handle_connection(self._lottery_state)
        self._clients.discard(client)

    def __accept_new_connection(self):
        """
        Accept new connections

        Function blocks until a connection to a client is made.
        Then connection created is printed and returned
        """

        # Connection arrived
        logging.info('action: accept_connections | result: in_progress')
        c, addr = self._server_socket.accept()
        logging.info(f'action: accept_connections | result: success | ip: {addr[0]}')
        return c
