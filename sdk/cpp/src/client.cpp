#include "lightlink/client.hpp"
#include <nats.hpp>
#include <sstream>
#include <random>
#include <chrono>
#include <fstream>

namespace lightlink {

// Implementation class (PIMPL pattern)
class ClientImpl {
public:
    std::string url;
    TLSConfig tls_config;
    std::unique_ptr<nats::Connection> nc;
    std::unique_ptr<nats::JetStream> js;
    std::map<std::string, std::unique_ptr<nats::Subscription>> subscriptions;
    std::map<std::string, std::unique_ptr<nats::KeyValue>> kv_stores;
    std::map<std::string, std::unique_ptr<nats::ObjectStore>> obj_stores;
    std::mutex mutex;

    // Generate UUID
    static std::string generate_uuid() {
        std::random_device rd;
        std::mt19937 gen(rd());
        std::uniform_int_distribution<> dis(0, 15);

        std::stringstream ss;
        ss << std::hex;
        for (int i = 0; i < 8; i++) ss << dis(gen);
        ss << "-";
        for (int i = 0; i < 4; i++) ss << dis(gen);
        ss << "-4";
        for (int i = 0; i < 3; i++) ss << dis(gen);
        ss << "-";
        ss << std::hex << (dis(gen) & 3 | 8);
        for (int i = 0; i < 3; i++) ss << dis(gen);
        ss << "-";
        for (int i = 0; i < 12; i++) ss << dis(gen);
        return ss.str();
    }

    // JSON helpers (simplified - in production use a proper JSON library)
    static std::string map_to_json(const std::map<std::string, std::string>& data) {
        std::stringstream ss;
        ss << "{";
        bool first = true;
        for (const auto& kv : data) {
            if (!first) ss << ",";
            ss << "\"" << kv.first << "\":\"" << kv.second << "\"";
            first = false;
        }
        ss << "}";
        return ss.str();
    }

    static std::map<std::string, std::string> json_to_map(const std::string& json) {
        std::map<std::string, std::string> result;
        // Simplified JSON parsing - in production use a proper JSON library
        // This is a placeholder for demonstration
        return result;
    }
};

// Client implementation

Client::Client(const std::string& url, const TLSConfig* tls_config)
    : impl_(std::make_unique<ClientImpl>()) {
    impl_->url = url;
    if (tls_config) {
        impl_->tls_config = *tls_config;
    }
}

Client::~Client() {
    close();
}

Client::Client(Client&&) noexcept = default;
Client& Client::operator=(Client&&) noexcept = default;

bool Client::connect() {
    std::lock_guard<std::mutex> lock(impl_->mutex);

    try {
        nats::Options options;
        options.setName("LightLink C++ Client");
        options.setReconnectWait(2000);
        options.setMaxReconnects(10);

        // Configure TLS if provided
        if (!impl_->tls_config.ca_file.empty()) {
            // Note: NATS C++ client TLS setup requires more configuration
            // This is a simplified version
        }

        impl_->nc = std::make_unique<nats::Connection>(nats::Connection::connect(impl_->url, options));
        impl_->js = std::make_unique<nats::JetStream>(impl_->nc->createJetStreamContext());
        return true;
    } catch (const std::exception& e) {
        return false;
    }
}

void Client::close() {
    std::lock_guard<std::mutex> lock(impl_->mutex);

    if (impl_->nc) {
        impl_->nc->close();
        impl_->nc.reset();
        impl_->js.reset();
        impl_->subscriptions.clear();
        impl_->kv_stores.clear();
        impl_->obj_stores.clear();
    }
}

bool Client::is_connected() const {
    std::lock_guard<std::mutex> lock(impl_->mutex);
    return impl_->nc && impl_->nc->isConnected();
}

std::map<std::string, std::string> Client::call(
    const std::string& service,
    const std::string& method,
    const std::map<std::string, std::string>& args,
    int timeout_ms) {

    std::map<std::string, std::string> result;

    std::string subject = "$SRV." + service + "." + method;

    // Build request JSON
    std::stringstream req_json;
    req_json << "{";
    req_json << "\"id\":\"" << impl_->generate_uuid() << "\",";
    req_json << "\"method\":\"" << method << "\",";
    req_json << "\"args\":{";
    bool first = true;
    for (const auto& arg : args) {
        if (!first) req_json << ",";
        req_json << "\"" << arg.first << "\":\"" << arg.second << "\"";
        first = false;
    }
    req_json << "}}";

    try {
        nats::Message msg = impl_->nc->request(subject, req_json.str(), timeout_ms);
        // Parse response (simplified)
        result["success"] = "true";
    } catch (const std::exception& e) {
        result["error"] = e.what();
    }

    return result;
}

void Client::call_async(
    const std::string& service,
    const std::string& method,
    const std::map<std::string, std::string>& args,
    RPCCallback callback,
    int timeout_ms) {
    // In a real implementation, this would use async/future
    // For now, call synchronously and invoke callback
    auto result = call(service, method, args, timeout_ms);
    callback(result, result.count("error") > 0 ? result["error"] : "");
}

bool Client::publish(const std::string& subject, const std::map<std::string, std::string>& data) {
    std::lock_guard<std::mutex> lock(impl_->mutex);

    if (!impl_->nc) return false;

    try {
        std::string json_data = impl_->map_to_json(data);
        impl_->nc->publish(subject, json_data);
        return true;
    } catch (const std::exception& e) {
        return false;
    }
}

std::string Client::subscribe(const std::string& subject, MessageHandler handler) {
    std::lock_guard<std::mutex> lock(impl_->mutex);

    if (!impl_->nc) return "";

    try {
        std::string sub_id = impl_->generate_uuid();

        auto sub = std::make_unique<nats::Subscription>(
            impl_->nc->subscribe(subject, [handler](nats::Message msg) {
                // Parse message data and invoke handler
                std::map<std::string, std::string> data;
                // Simplified parsing - in production use proper JSON library
                handler(data);
            })
        );

        impl_->subscriptions[sub_id] = std::move(sub);
        return sub_id;
    } catch (const std::exception& e) {
        return "";
    }
}

void Client::unsubscribe(const std::string& subscription_id) {
    std::lock_guard<std::mutex> lock(impl_->mutex);

    auto it = impl_->subscriptions.find(subscription_id);
    if (it != impl_->subscriptions.end()) {
        it->second->unsubscribe();
        impl_->subscriptions.erase(it);
    }
}

bool Client::set_state(const std::string& key, const std::map<std::string, std::string>& value) {
    std::lock_guard<std::mutex> lock(impl_->mutex);

    if (!impl_->js) return false;

    try {
        // Get or create KV store
        if (impl_->kv_stores.find("light_link_state") == impl_->kv_stores.end()) {
            try {
                impl_->kv_stores["light_link_state"] = std::make_unique<nats::KeyValue>(
                    impl_->js->getKeyValue("light_link_state")
                );
            } catch (...) {
                impl_->kv_stores["light_link_state"] = std::make_unique<nats::KeyValue>(
                    impl_->js->createKeyValue("light_link_state")
                );
            }
        }

        std::string json_value = impl_->map_to_json(value);
        impl_->kv_stores["light_link_state"]->put(key, json_value);
        return true;
    } catch (const std::exception& e) {
        return false;
    }
}

std::map<std::string, std::string> Client::get_state(const std::string& key) {
    std::lock_guard<std::mutex> lock(impl_->mutex);

    std::map<std::string, std::string> result;

    if (!impl_->js) return result;

    try {
        if (impl_->kv_stores.find("light_link_state") == impl_->kv_stores.end()) {
            impl_->kv_stores["light_link_state"] = std::make_unique<nats::KeyValue>(
                impl_->js->getKeyValue("light_link_state")
            );
        }

        nats::KeyValueEntry entry = impl_->kv_stores["light_link_state"]->get(key);
        // Parse JSON value (simplified)
        return impl_->json_to_map(entry.getValue());
    } catch (const std::exception& e) {
        result["error"] = e.what();
    }

    return result;
}

std::string Client::watch_state(const std::string& key, MessageHandler handler) {
    std::lock_guard<std::mutex> lock(impl_->mutex);

    if (!impl_->js) return "";

    try {
        std::string watch_id = impl_->generate_uuid();

        // Get or create KV store
        if (impl_->kv_stores.find("light_link_state") == impl_->kv_stores.end()) {
            try {
                impl_->kv_stores["light_link_state"] = std::make_unique<nats::KeyValue>(
                    impl_->js->getKeyValue("light_link_state")
                );
            } catch (...) {
                impl_->kv_stores["light_link_state"] = std::make_unique<nats::KeyValue>(
                    impl_->js->createKeyValue("light_link_state")
                );
            }
        }

        // Create watcher (implementation depends on NATS C++ client API)
        // This is a placeholder for the actual implementation

        return watch_id;
    } catch (const std::exception& e) {
        return "";
    }
}

void Client::unwatch_state(const std::string& watch_id) {
    // Stop watching state changes
}

std::string Client::upload_file(const std::string& file_path, const std::string& remote_name) {
    std::lock_guard<std::mutex> lock(impl_->mutex);

    if (!impl_->js) return "";

    try {
        // Get or create object store
        if (impl_->obj_stores.find("light_link_files") == impl_->obj_stores.end()) {
            try {
                impl_->obj_stores["light_link_files"] = std::make_unique<nats::ObjectStore>(
                    impl_->js->getObjectStore("light_link_files")
                );
            } catch (...) {
                impl_->obj_stores["light_link_files"] = std::make_unique<nats::ObjectStore>(
                    impl_->js->createObjectStore("light_link_files")
                );
            }
        }

        std::string file_id = impl_->generate_uuid();

        // Read file and upload in chunks
        std::ifstream file(file_path, std::ios::binary);
        if (!file) return "";

        const size_t chunk_size = 1024 * 1024; // 1MB chunks
        std::vector<char> buffer(chunk_size);
        int chunk_num = 0;

        while (file) {
            file.read(buffer.data(), chunk_size);
            size_t bytes_read = file.gcount();

            if (bytes_read > 0) {
                std::string chunk_key = file_id + "_" + std::to_string(chunk_num);
                std::string chunk_data(buffer.data(), bytes_read);

                // Upload chunk (implementation depends on NATS C++ client API)
                chunk_num++;
            }
        }

        return file_id;
    } catch (const std::exception& e) {
        return "";
    }
}

bool Client::download_file(const std::string& file_id, const std::string& local_path) {
    std::lock_guard<std::mutex> lock(impl_->mutex);

    if (!impl_->js) return false;

    try {
        // Get object store
        if (impl_->obj_stores.find("light_link_files") == impl_->obj_stores.end()) {
            impl_->obj_stores["light_link_files"] = std::make_unique<nats::ObjectStore>(
                impl_->js->getObjectStore("light_link_files")
            );
        }

        // Download chunks and write to file
        std::ofstream file(local_path, std::ios::binary);
        if (!file) return false;

        int chunk_num = 0;
        while (true) {
            std::string chunk_key = file_id + "_" + std::to_string(chunk_num);

            try {
                // Get chunk (implementation depends on NATS C++ client API)
                chunk_num++;
            } catch (...) {
                break; // No more chunks
            }
        }

        return true;
    } catch (const std::exception& e) {
        return false;
    }
}

} // namespace lightlink
