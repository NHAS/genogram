#include "node_editor.h"
#include <imnodes.h>
#include <imgui.h>
#include <misc/cpp/imgui_stdlib.h>
#include <SDL2/SDL_scancode.h>

#include <string>
#include <algorithm>
#include <vector>
#include <iostream>

namespace geneogram
{
    namespace
    {
        struct Node
        {

            int id;
            std::string date_of_birth;
            std::string name;

            Node(const int i, const std::string dob, const std::string n) : id(i), date_of_birth(dob), name(n) {}
        };

        struct Link
        {
            int id;
            int start_attr, end_attr;
        };

        struct Editor
        {
            ImNodesEditorContext *context = nullptr;
            std::vector<Node> nodes;
            std::vector<Link> links;
            int current_id = 0;

            bool newPersonMenu;
            std::string inputName, dob;
            ImVec2 MouseOnOpening;
        };

        void newPersonMenu(Editor &editor)
        {
            ImGui::SetNextWindowPos(editor.MouseOnOpening, ImGuiCond_FirstUseEver);
            if (ImGui::Begin("New Person", &editor.newPersonMenu, ImGuiWindowFlags_AlwaysAutoResize))
            {
                ImGui::Text("Enter details of new person");

                bool reclaim_focus = false;
                ImGuiInputTextFlags input_text_flags = ImGuiInputTextFlags_EnterReturnsTrue | ImGuiInputTextFlags_CallbackCompletion | ImGuiInputTextFlags_CallbackHistory;
                if (ImGui::InputText(
                        "Name##", &editor.inputName, input_text_flags))
                {
                    reclaim_focus = true;
                }

                reclaim_focus = false;
                if (ImGui::InputText(
                        "DOB##", &editor.dob, input_text_flags))
                {
                    reclaim_focus = true;
                }
                // Auto-focus on window apparition
                ImGui::SetItemDefaultFocus();
                if (reclaim_focus)
                    ImGui::SetKeyboardFocusHere(-1); // Auto focus previous widget

                if (ImGui::Button("Done##"))
                {
                    const int node_id = ++editor.current_id;
                    ImNodes::SetNodeScreenSpacePos(node_id, editor.MouseOnOpening);
                    ImNodes::SnapNodeToGrid(node_id);
                    editor.nodes.push_back(Node(node_id, editor.dob, editor.inputName));
                    editor.inputName = "";
                    editor.dob = "";
                    editor.newPersonMenu = false;
                }

                ImGui::End();
            }
        }

        void show_editor(const char *editor_name, Editor &editor)
        {
#ifdef IMGUI_HAS_VIEWPORT
            ImGuiViewport *viewport = ImGui::GetMainViewport();
            ImGui::SetNextWindowPos(viewport->GetWorkPos());
            ImGui::SetNextWindowSize(viewport->GetWorkSize());
            ImGui::SetNextWindowViewport(viewport->ID);
#else
            ImGui::SetNextWindowPos(ImVec2(0.0f, 0.0f));
            ImGui::SetNextWindowSize(ImGui::GetIO().DisplaySize);
#endif

            ImNodes::EditorContextSet(editor.context);

            ImGui::Begin(
                editor_name,
                nullptr,
                ImGuiWindowFlags_NoResize | ImGuiWindowFlags_HorizontalScrollbar |
                    ImGuiWindowFlags_NoDecoration | ImGuiWindowFlags_NoTitleBar | ImGuiWindowFlags_NoBringToFrontOnFocus);
            ImGui::TextUnformatted("");

            ImNodes::BeginNodeEditor();

            const bool open_popup = ImGui::IsWindowFocused(ImGuiFocusedFlags_RootAndChildWindows) &&
                                    ImNodes::IsEditorHovered() &&
                                    ImGui::IsMouseClicked(ImGuiMouseButton_Right);

            ImGui::PushStyleVar(ImGuiStyleVar_WindowPadding, ImVec2(8.f, 8.f));
            if (!ImGui::IsAnyItemHovered() && open_popup)
            {
                ImGui::OpenPopup("add node");
            }

            if (ImGui::BeginPopup("add node"))
            {
                editor.MouseOnOpening = ImGui::GetMousePosOnOpeningCurrentPopup();
                if (ImGui::MenuItem("add"))
                {
                    editor.newPersonMenu = true;
                }

                ImGui::EndPopup();
            }

            if (editor.newPersonMenu)
            {
                newPersonMenu(editor);
            }

            ImGui::PopStyleVar();

            for (Node &node : editor.nodes)
            {
                ImNodes::BeginNode(node.id);

                ImNodes::BeginNodeTitleBar();
                ImGui::TextUnformatted(node.name.c_str());
                ImNodes::EndNodeTitleBar();

                ImNodes::BeginInputAttribute(node.id << 8);
                ImGui::TextUnformatted("parents");
                ImNodes::EndInputAttribute();

                ImNodes::BeginStaticAttribute(node.id << 16);
                ImGui::TextUnformatted(node.date_of_birth.c_str());
                ImNodes::EndStaticAttribute();

                ImNodes::BeginOutputAttribute(node.id << 24);
                ImGui::TextUnformatted("children");
                ImNodes::EndOutputAttribute();

                ImNodes::EndNode();
            }

            for (const Link &link : editor.links)
            {
                ImNodes::Link(link.id, link.start_attr, link.end_attr);
            }

            ImNodes::MiniMap(0.2f, ImNodesMiniMapLocation_BottomRight);
            ImNodes::EndNodeEditor();

            {
                Link link;
                if (ImNodes::IsLinkCreated(&link.start_attr, &link.end_attr))
                {
                    link.id = ++editor.current_id;
                    editor.links.push_back(link);
                }
            }

            {
                int link_id;
                if (ImNodes::IsLinkDestroyed(&link_id))
                {
                    auto iter = std::find_if(
                        editor.links.begin(), editor.links.end(), [link_id](const Link &link) -> bool
                        { return link.id == link_id; });
                    assert(iter != editor.links.end());
                    editor.links.erase(iter);
                }
            }

            {
                if (ImGui::IsKeyReleased(ImGuiKey_Delete))
                {
                    const int num_selected_links = ImNodes::NumSelectedLinks();

                    if (num_selected_links > 0)
                    {
                        static std::vector<int> selected_links;
                        selected_links.resize(static_cast<size_t>(num_selected_links));

                        ImNodes::GetSelectedLinks(selected_links.data());
                        for (const int edge_id : selected_links)
                        {
                            // O(n^2) babyyyyyy
                            auto iter = std::find_if(
                                editor.links.begin(), editor.links.end(), [edge_id](const Link &link) -> bool
                                { return link.id == edge_id; });

                            editor.links.erase(iter);
                        }
                    }

                    const int num_selected_nodes = ImNodes::NumSelectedNodes();
                    if (num_selected_nodes > 0)
                    {
                        static std::vector<int> selected_nodes;
                        selected_nodes.resize(static_cast<size_t>(num_selected_nodes));

                        ImNodes::GetSelectedNodes(selected_nodes.data());
                        for (const int node_id : selected_nodes)
                        {
                            // O(n^2) babyyyyyy
                            auto iter = std::find_if(
                                editor.nodes.begin(), editor.nodes.end(), [node_id](const Node &node) -> bool
                                { return node.id == node_id; });

                            editor.nodes.erase(iter);
                        }
                    }
                }
            }

            ImGui::End();
        }

        Editor editor1;

    } // namespace

    void NodeEditorInitialize()
    {
        editor1.context = ImNodes::EditorContextCreate();
        ImNodes::PushAttributeFlag(ImNodesAttributeFlags_EnableLinkDetachWithDragClick);

        ImNodesIO &io = ImNodes::GetIO();
        io.LinkDetachWithModifierClick.Modifier = &ImGui::GetIO().KeyCtrl;
        io.MultipleSelectModifier.Modifier = &ImGui::GetIO().KeyCtrl;

        ImNodesStyle &style = ImNodes::GetStyle();
        style.Flags |= ImNodesStyleFlags_GridLinesPrimary | ImNodesStyleFlags_GridSnapping;
    }

    void NodeEditorShow() { show_editor("editor1", editor1); }

    void NodeEditorShutdown()
    {
        ImNodes::PopAttributeFlag();
        ImNodes::EditorContextFree(editor1.context);
    }

} // namespace example
