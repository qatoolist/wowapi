#!/usr/bin/env python3
import json
import sys
import datetime

def generate_report(json_data):
    results = {}
    for line in json_data.strip().split('\n'):
        try:
            entry = json.loads(line)
        except json.JSONDecodeError:
            continue
        
        pkg = entry.get('Package', 'unknown')
        if pkg not in results:
            results[pkg] = {'Action': entry.get('Action'), 'Tests': {}}
            
        test = entry.get('Test')
        if test:
            if test not in results[pkg]['Tests']:
                results[pkg]['Tests'][test] = {'Output': []}
            if entry.get('Action') == 'output':
                results[pkg]['Tests'][test]['Output'].append(entry.get('Output', ''))
            else:
                results[pkg]['Tests'][test]['Status'] = entry.get('Action')

    html = f"""
    <html>
    <head>
    <title>Test Execution Report - {datetime.datetime.now().strftime('%Y-%m-%d %H:%M:%S')}</title>
    <style>
        body {{ font-family: sans-serif; padding: 20px; }}
        .pass {{ color: green; }}
        .fail {{ color: red; }}
        table {{ width: 100%; border-collapse: collapse; }}
        th, td {{ border: 1px solid #ccc; padding: 8px; text-align: left; }}
    </style>
    </head>
    <body>
    <h1>Test Execution Report</h1>
    <table>
        <tr><th>Package</th><th>Test</th><th>Status</th></tr>
    """
    
    for pkg, data in results.items():
        for test, info in data['Tests'].items():
            status = info.get('Status', 'unknown')
            cls = 'pass' if status == 'pass' else 'fail'
            html += f"<tr><td>{pkg}</td><td>{test}</td><td class='{cls}'>{status}</td></tr>"
            
    html += "</table></body></html>"
    return html

if __name__ == "__main__":
    data = sys.stdin.read()
    print(generate_report(data))
