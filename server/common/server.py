import socket
import logging
import threading
from .client import Client
from .utils import store_bets, load_bets, has_won

class LotteryState:
    def __init__(self, expected_agencies):
        self._expected_agencies = expected_agencies
        self._agencies_finish = 0
        self._lock = threading.Lock()
    
    def agency_finished(self):
        with self._lock:
            self._agencies_finish += 1
            if self._agencies_finish == self._expected_agencies:
                logging.info("action: sorteo | result: success")

    def winners_ready(self):
        with self._lock:
            return self._agencies_finish >= self._expected_agencies

class BetsRepository:
    def __init__(self):
        self._lock = threading.Lock()
    
    def store_bets(self, bets):
        with self._lock:
            store_bets(bets)
    
    def load_bets(self):
        with self._lock:
            return list(load_bets())
    
    def has_won(self, bet):
        return has_won(bet)

class SetMonitor:
    def __init__(self):
        self._set = set()
        self._lock = threading.Lock()
    
    def add(self, item):
        with self._lock:
            self._set.add(item)
    
    def discard(self, item):
        with self._lock:
            self._set.discard(item)
    
    def copy(self):
        with self._lock:
            return set(self._set)

class Server:
    def __init__(self, port, listen_backlog, expected_agencies):
        # Initialize server socket
        self._server_socket = socket.socket(socket.AF_INET, socket.SOCK_STREAM)
        self._server_socket.bind(('', port))
        self._server_socket.listen(listen_backlog)
        self._running = True
        self._clients = SetMonitor()
        self._lottery_state = LotteryState(expected_agencies)
        self._threads = SetMonitor()
        self._bets_repo = BetsRepository()

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
        
        for thread in self._threads.copy():
            thread.join()
        
        self._threads = SetMonitor()
        
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
                    client = Client(client_sock, self._bets_repo)
                    self._clients.add(client)
                    thread = threading.Thread(
                        target=self.__handle_client_connection,
                        args=(client,)
                    )
                    self._threads.add(thread)
                    thread.start()
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
        self._threads.discard(threading.current_thread())

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
