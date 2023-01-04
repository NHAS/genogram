cmake_minimum_required(VERSION 3.15)

project(genogram)

set(CMAKE_CXX_STANDARD 14)
set(CMAKE_CXX_STANDARD_REQUIRED True)

find_package(SDL2 CONFIG REQUIRED)

include_directories(genogram ${SDL2_INCLUDE_DIRS})
message("SDL2 Include path ${SDL2_INCLUDE_DIRS}")


if(MSVC)
    add_compile_definitions(SDL_MAIN_HANDLED)
    add_compile_options(/W4 /WX)
else()
    add_compile_options(-Wall -Wextra -Wpedantic -Werror)
endif()


set(IMGUI_DIR ${CMAKE_CURRENT_SOURCE_DIR}/imgui)
add_library(IMGUI STATIC)

target_sources( IMGUI
                PRIVATE
                    ${IMGUI_DIR}/misc/cpp/imgui_stdlib.cpp
                    ${IMGUI_DIR}/imgui_demo.cpp
                    ${IMGUI_DIR}/imgui_draw.cpp
                    ${IMGUI_DIR}/imgui_tables.cpp
                    ${IMGUI_DIR}/imgui_widgets.cpp
                    ${IMGUI_DIR}/imgui.cpp

                PRIVATE
                    ${IMGUI_DIR}/backends/imgui_impl_opengl3.cpp
                    ${IMGUI_DIR}/backends/imgui_impl_sdl.cpp
                )

target_include_directories( IMGUI
                            PUBLIC ${IMGUI_DIR}
                            PUBLIC ${IMGUI_DIR}/backends
                            PUBLIC ${IMGUI_DIR}/misc/cpp
                          )


# Imnodes

add_library(imnodes imnodes)
target_sources(imnodes PRIVATE
    ${CMAKE_SOURCE_DIR}/imnodes/imnodes.h
    ${CMAKE_SOURCE_DIR}/imnodes/imnodes_internal.h
    ${CMAKE_SOURCE_DIR}/imnodes/imnodes.cpp)
target_include_directories(imnodes PUBLIC ${CMAKE_SOURCE_DIR}/imnodes)
target_link_libraries(imnodes PUBLIC IMGUI)

# Example projects

add_executable(genogram
    ${CMAKE_SOURCE_DIR}/main.cpp
    ${CMAKE_SOURCE_DIR}/renderer.cpp)
target_link_libraries(genogram imnodes SDL2::SDL2)
if (APPLE)
    target_link_libraries(genogram "-framework OpenGL")
elseif(MSVC)
    target_link_libraries(genogram "opengl32")
else()
    target_link_libraries(genogram X11 Xext GL)
endif()