import os
import socket
import threading
import time

import paramiko


SSH_HOST = os.environ.get("CLOUD_DB_SSH_HOST", "43.165.167.165")
SSH_USER = os.environ.get("CLOUD_DB_SSH_USER", "ubuntu")
SSH_PASSWORD = os.environ.get("CLOUD_DB_SSH_PASSWORD")
LOCAL_HOST = os.environ.get("CLOUD_DB_LOCAL_HOST", "127.0.0.1")
LOCAL_PORT = int(os.environ.get("CLOUD_DB_LOCAL_PORT", "5433"))
REMOTE_HOST = os.environ.get("CLOUD_DB_REMOTE_HOST", "127.0.0.1")
REMOTE_PORT = int(os.environ.get("CLOUD_DB_REMOTE_PORT", "5432"))


def pipe(src, dst):
    try:
        while True:
            data = src.recv(32768)
            if not data:
                break
            dst.sendall(data)
    except Exception:
        pass
    finally:
        for sock in (src, dst):
            try:
                sock.close()
            except Exception:
                pass


def handle_client(client_sock, transport):
    try:
        channel = transport.open_channel(
            "direct-tcpip",
            (REMOTE_HOST, REMOTE_PORT),
            client_sock.getsockname(),
        )
    except Exception:
        client_sock.close()
        return

    threading.Thread(target=pipe, args=(client_sock, channel), daemon=True).start()
    threading.Thread(target=pipe, args=(channel, client_sock), daemon=True).start()


def run_once():
    ssh = paramiko.SSHClient()
    ssh.set_missing_host_key_policy(paramiko.AutoAddPolicy())
    ssh.connect(
        SSH_HOST,
        port=22,
        username=SSH_USER,
        password=SSH_PASSWORD,
        timeout=20,
        banner_timeout=20,
        auth_timeout=20,
    )

    transport = ssh.get_transport()
    transport.set_keepalive(30)

    server = socket.socket(socket.AF_INET, socket.SOCK_STREAM)
    server.setsockopt(socket.SOL_SOCKET, socket.SO_REUSEADDR, 1)
    server.bind((LOCAL_HOST, LOCAL_PORT))
    server.listen(100)

    try:
        while transport and transport.is_active():
            client_sock, _ = server.accept()
            threading.Thread(
                target=handle_client,
                args=(client_sock, transport),
                daemon=True,
            ).start()
    finally:
        server.close()
        ssh.close()


def main():
    if not SSH_PASSWORD:
        raise SystemExit("CLOUD_DB_SSH_PASSWORD is required")

    while True:
        try:
            run_once()
        except OSError as exc:
            if "10048" in str(exc) or "Address already in use" in str(exc):
                raise
            time.sleep(5)
        except Exception:
            time.sleep(5)


if __name__ == "__main__":
    main()
