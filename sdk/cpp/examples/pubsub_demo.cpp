#include "lightlink/client.hpp"
#include <iostream>
#include <thread>
#include <chrono>
#include <atomic>

int main() {
    std::cout << "=== C++ SDK Pub/Sub Demo ===" << std::endl;

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

    // Subscribe
    std::cout << "\n[2/2] Testing pub/sub..." << std::endl;
    std::atomic<int> received_count(0);

    std::string sub_id = client.subscribe("test.cpp", [&received_count](const auto& data) {
        std::cout << "  Received message #" << (received_count + 1) << std::endl;
        received_count++;
    });

    if (sub_id.empty()) {
        std::cerr << "Failed to subscribe" << std::endl;
        return 1;
    }

    // Publish messages
    for (int i = 0; i < 3; i++) {
        std::map<std::string, std::string> data;
        data["id"] = std::to_string(i);
        data["msg"] = "Hello from C++ #" + std::to_string(i);

        if (client.publish("test.cpp", data)) {
            std::cout << "  Published message #" << (i + 1) << std::endl;
        }
    }

    // Wait for messages
    std::this_thread::sleep_for(std::chrono::seconds(1));

    // Unsubscribe
    client.unsubscribe(sub_id);

    std::cout << "  Total received: " << received_count << " messages" << std::endl;

    // Cleanup
    client.close();

    std::cout << "\n=== C++ SDK Pub/Sub Demo Complete ===" << std::endl;
    return 0;
}
