#include "lightlink/client.hpp"
#include <iostream>
#include <thread>
#include <chrono>

int main() {
    std::cout << "=== C++ SDK RPC Demo ===" << std::endl;

    // Get NATS URL from environment or use default
    const char* nats_url_env = std::getenv("NATS_URL");
    std::string nats_url = nats_url_env ? nats_url_env : "nats://172.18.200.47:4222";

    std::cout << "NATS URL: " << nats_url << std::endl;

    // Create client
    lightlink::Client client(nats_url);

    // Connect
    std::cout << "\n[1/2] Connecting to NATS..." << std::endl;
    if (!client.connect()) {
        std::cerr << "Failed to connect" << std::endl;
        return 1;
    }
    std::cout << "Connected successfully!" << std::endl;

    // RPC call
    std::cout << "\n[2/2] Testing RPC call..." << std::endl;
    std::map<std::string, std::string> args;
    args["a"] = "100";
    args["b"] = "200";

    auto result = client.call("demo-service", "add", args);

    if (result.count("error")) {
        std::cout << "RPC Error: " << result["error"] << std::endl;
    } else {
        std::cout << "RPC Result: sum=" << (result.count("sum") ? result["sum"] : "N/A") << std::endl;
    }

    // Cleanup
    client.close();

    std::cout << "\n=== C++ SDK RPC Demo Complete ===" << std::endl;
    return 0;
}
