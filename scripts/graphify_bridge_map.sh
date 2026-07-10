#!/bin/sh
set -eu

mode="${1:-node}"
graph_path="${GRAPHIFY_GRAPH_PATH:-graphify-out/graph.json}"

if [ ! -f "$graph_path" ]; then
  echo "graphify bridge map: missing graph file: $graph_path" >&2
  exit 2
fi

python3 - "$mode" "$graph_path" <<'PY'
import json
import sys
from collections import defaultdict

mode = sys.argv[1]
graph_path = sys.argv[2]

with open(graph_path, "r", encoding="utf-8") as fh:
    graph = json.load(fh)

nodes = {n["id"]: n for n in graph.get("nodes", [])}
links = graph.get("links") or graph.get("edges") or []

neighbors = defaultdict(set)
cross_counts = defaultdict(lambda: defaultdict(int))
bridge_rows = []

def community_name(node):
    return node.get("community_name") or str(node.get("community") or "")

for link in links:
    src = link.get("source")
    dst = link.get("target")
    if src not in nodes or dst not in nodes:
        continue
    neighbors[src].add(dst)
    neighbors[dst].add(src)

    cs = community_name(nodes[src])
    ct = community_name(nodes[dst])
    if cs and ct and cs != ct:
        cross_counts[cs][ct] += 1
        cross_counts[ct][cs] += 1

for nid, node in nodes.items():
    comm = community_name(node)
    if not comm:
        continue
    peer_comms = sorted({
        community_name(nodes[other])
        for other in neighbors.get(nid, set())
        if community_name(nodes[other]) and community_name(nodes[other]) != comm
    })
    if peer_comms:
        bridge_rows.append((
            comm,
            node.get("label", ""),
            len(peer_comms),
            ", ".join(peer_comms),
            len(neighbors.get(nid, set())),
            node.get("source_file", ""),
        ))

if mode == "node":
    print("community\tlabel\tcross_communities\tpeer_communities\tdegree\tsource_file")
    for row in sorted(bridge_rows, key=lambda r: (-r[2], -r[4], r[0], r[1])):
        print("\t".join(map(str, row)))
elif mode == "community":
    print("community\tpeer_community\tedge_count")
    for comm in sorted(cross_counts):
        peers = cross_counts[comm]
        for peer, count in sorted(peers.items(), key=lambda kv: (-kv[1], kv[0])):
            if comm < peer:
                print(f"{comm}\t{peer}\t{count}")
else:
    print("usage: graphify_bridge_map.sh [node|community]" , file=sys.stderr)
    sys.exit(2)
PY
