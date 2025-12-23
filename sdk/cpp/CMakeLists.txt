cmake_minimum_required(VERSION 3.15)
project(lightlink_cpp VERSION 0.1.0 LANGUAGES CXX)

set(CMAKE_CXX_STANDARD 17)
set(CMAKE_CXX_STANDARD_REQUIRED ON)

# Options
option(BUILD_EXAMPLES "Build examples" ON)
option(BUILD_TESTS "Build tests" ON)

# Find NATS C++ client
find_package(natscpp CONFIG REQUIRED)

# Library
add_library(lightlink STATIC
    src/client.cpp
)

target_include_directories(lightlink PUBLIC
    $<BUILD_INTERFACE:${CMAKE_CURRENT_SOURCE_DIR}/include>
    $<INSTALL_INTERFACE:include>
)

target_link_libraries(lightlink PUBLIC natscpp::natscpp)

# Examples
if(BUILD_EXAMPLES)
    add_executable(rpc_example examples/rpc_demo.cpp)
    target_link_libraries(rpc_example PRIVATE lightlink)

    add_executable(pubsub_example examples/pubsub_demo.cpp)
    target_link_libraries(pubsub_example PRIVATE lightlink)
endif()
