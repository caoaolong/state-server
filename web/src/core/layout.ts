import dagre from "@dagrejs/dagre";
import { Position, useVueFlow, type Edge, type Node } from "@vue-flow/core";
import { ref } from "vue";

export function useLayout() {
  const { findNode } = useVueFlow();

  const graph = ref(new dagre.graphlib.Graph());

  const previousDirection = ref("LR");

  function layout(nodes: Node[], edges: Edge[], direction: string) {
    const dagreGraph = new dagre.graphlib.Graph();

    graph.value = dagreGraph;

    dagreGraph.setDefaultEdgeLabel(() => ({}));

    const isHorizontal = direction === "LR";
    dagreGraph.setGraph({ rankdir: direction });

    previousDirection.value = direction;

    for (const node of nodes) {
      const graphNode = findNode(node.id);
      if (!graphNode) continue;
      dagreGraph.setNode(node.id, {
        width: graphNode.dimensions.width || 150,
        height: graphNode.dimensions.height || 50,
      });
    }

    for (const edge of edges) {
      dagreGraph.setEdge(edge.source, edge.target);
    }

    dagre.layout(dagreGraph);

    return nodes.map((node) => {
      const np = dagreGraph.node(node.id);
      const position = np
        ? { x: np.x, y: np.y }
        : (node.position ?? { x: 0, y: 0 });
      return {
        ...node,
        targetPosition: isHorizontal ? Position.Left : Position.Top,
        sourcePosition: isHorizontal ? Position.Right : Position.Bottom,
        position,
      };
    });
  }

  return { graph, layout, previousDirection };
}
