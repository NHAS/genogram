#include "node_editor.h"
#include <imnodes.h>
#include <imgui.h>
#include <SDL2/SDL_scancode.h>

#include <string>
#include <algorithm>
#include <vector>
#include <iostream>

namespace example
{
    namespace
    {
        struct Node
        {
            std::string name;
            int id;
            float value;

            Node(const int i, const float v) : id(i), value(v) {}
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
        };

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
                    ImGuiWindowFlags_NoDecoration | ImGuiWindowFlags_NoTitleBar);
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
                const ImVec2 click_pos = ImGui::GetMousePosOnOpeningCurrentPopup();

                if (ImGui::MenuItem("add"))
                {
                    const int node_id = ++editor.current_id;
                    ImNodes::SetNodeScreenSpacePos(node_id, click_pos);
                    ImNodes::SnapNodeToGrid(node_id);
                    editor.nodes.push_back(Node(node_id, 0.f));
                }

                ImGui::EndPopup();
            }

            ImGui::PopStyleVar();

            for (Node &node : editor.nodes)
            {
                ImNodes::BeginNode(node.id);

                ImNodes::BeginNodeTitleBar();
                ImGui::TextUnformatted("node");
                ImNodes::EndNodeTitleBar();

                ImNodes::BeginInputAttribute(node.id << 8);
                ImGui::TextUnformatted("input");
                ImNodes::EndInputAttribute();

                ImNodes::BeginStaticAttribute(node.id << 16);
                ImGui::PushItemWidth(120.0f);
                ImGui::DragFloat("value", &node.value, 0.01f);
                ImGui::PopItemWidth();
                ImNodes::EndStaticAttribute();

                ImNodes::BeginOutputAttribute(node.id << 24);
                const float text_width = ImGui::CalcTextSize("output").x;
                ImGui::Indent(120.f + ImGui::CalcTextSize("value").x - text_width);
                ImGui::TextUnformatted("output");
                ImNodes::EndOutputAttribute();

                ImNodes::EndNode();
            }

            for (const Link &link : editor.links)
            {
                ImNodes::Link(link.id, link.start_attr, link.end_attr);
            }

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
