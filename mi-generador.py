import sys

def build_compose(n_clients):
    lines = [
        "name: tp0",
        "services:",
        "  server:",
        "    container_name: server",
        "    image: server:latest",
        "    entrypoint: python3 /main.py",
        "    environment:",
        "      - PYTHONUNBUFFERED=1",
        "      - LOGGING_LEVEL=DEBUG",
        "    networks:",
        "      - testing_net",
        "",
    ]

    for i in range(1, n_clients + 1):
        lines.extend([
            f"  client{i}:",
            f"    container_name: client{i}",
            "    image: client:latest",
            "    entrypoint: /client",
            "    environment:",
            f"      - CLI_ID={i}",
            "      - CLI_LOG_LEVEL=DEBUG",
            "    networks:",
            "      - testing_net",
            "    depends_on:",
            "      - server",
            "",
        ])

    lines.extend([
        "networks:",
        "  testing_net:",
        "    ipam:",
        "      driver: default",
        "      config:",
        "        - subnet: 172.25.125.0/24",
        "",
    ])

    return "\n".join(lines)

def main():
    if len(sys.argv) != 3:
        print("Usage: python3 mi-generador.py <output_file> <n_clients>")
        return 1

    output_file = sys.argv[1]
    try:
        n_clients = int(sys.argv[2])
    except ValueError:
        print("Error: n_clients must be an integer.")
        return 1
    
    content = build_compose(n_clients)

    with open(output_file, 'w') as f:
        written = f.write(content)
    
    if written != len(content):
        print("Error: short write")
        return 1

    print(f"Generated {output_file} with {n_clients} clients.")
    return 0

if __name__ == "__main__":
    sys.exit(main())
