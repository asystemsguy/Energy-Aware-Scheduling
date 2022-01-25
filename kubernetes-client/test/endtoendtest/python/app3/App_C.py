
#!/usr/bin/env python

from http.server import HTTPServer, BaseHTTPRequestHandler
from socketserver import ThreadingMixIn
import http.server
import os
import logging
import random
import hashlib
import time
import sys
import requests
import socketserver

logging.basicConfig(level=logging.DEBUG,
                    format='%(asctime)s %(levelname)-8s %(message)s',
                    datefmt='%a, %d %b %Y %H:%M:%S',
                    filename='/temp/myapp.log',
                    filemode='w')

# configuration variables
message_size = 12
talk_with_a = "10.96.0.11:8090/c"
talk_with_b = "10.96.0.12:8091/c"
talk_with_c = "10.96.0.13:8092/c"
talk_with_d = "10.96.0.14:8093/c"
talk_with_e = "10.96.0.15:8094/c"


def get_messge(length=12):
    message = [0] * length
    return message

def send_message(talk_with):

    sys.stdout.write("Started requests form C to " +talk_with+ '\n')

    # Replace with the correct URL
    url = "http://"+talk_with

    Response = requests.get(url)

    print (Response.status_code)


class KodeFunHTTPRequestHandler(BaseHTTPRequestHandler):

    def do_GET(self):
        os.system('python /server/CPULoadGenerator.py -l 0.003 -d 3 -c 0')
        
        if self.path == '/a':
             
            send_message(talk_with_a)
            sys.stdout.write("TEST A->A" + '\n')
            return 

        if self.path == '/b':
             
            send_message(talk_with_e)
            sys.stdout.write("message C->E" + '\n')
            self.send_response_t(7000000)
            sys.stdout.write("Response B<-C" + '\n')
            return 

        if self.path == '/c':

            send_message(talk_with_e)
            sys.stdout.write("TEST C->E" + '\n')
            return

        if self.path == '/d':

            send_message(talk_with_d)
            sys.stdout.write("TEST A->D" + '\n')
            return

        if self.path == '/e':
            
            send_message(talk_with_e)
            sys.stdout.write("TEST A->E" + '\n')
            return  

    def send_response_t(self, size):

        return_str = str(get_messge(size))

        self.send_response(200)
        self.send_header('Content-type','text/html')
        self.end_headers()
        # Send the html message
        self.wfile.write(return_str.encode())
        sys.stdout.write("message sent and recived" + '\n')

class ThreadedHTTPServer(socketserver.ThreadingMixIn, http.server.HTTPServer):
    daemon_threads = True
    
def run():
    logging.info('http server is starting...')

    server_address = ('', 8090)
    httpd = ThreadedHTTPServer(server_address, KodeFunHTTPRequestHandler)
    logging.info('http server is running...')
    httpd.serve_forever()

if __name__ == '__main__':
    run()
