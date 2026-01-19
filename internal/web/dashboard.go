/*
Copyright 2025.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package web

const dashboardHTML = `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>KubeSleuth Dashboard</title>
    <style>
        * {
            margin: 0;
            padding: 0;
            box-sizing: border-box;
        }
        body {
            font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, Oxygen, Ubuntu, Cantarell, sans-serif;
            background: #f5f5f5;
            color: #333;
            padding: 20px;
        }
        .container {
            max-width: 1400px;
            margin: 0 auto;
            background: white;
            border-radius: 8px;
            box-shadow: 0 2px 4px rgba(0,0,0,0.1);
            padding: 24px;
        }
        h1 {
            color: #1a1a1a;
            margin-bottom: 8px;
            font-size: 28px;
        }
        .subtitle {
            color: #666;
            margin-bottom: 24px;
            font-size: 14px;
        }
        .stats {
            display: flex;
            gap: 16px;
            margin-bottom: 24px;
        }
        .stat-card {
            flex: 1;
            background: #f8f9fa;
            padding: 16px;
            border-radius: 6px;
            border-left: 4px solid #007bff;
        }
        .stat-label {
            font-size: 12px;
            color: #666;
            text-transform: uppercase;
            margin-bottom: 4px;
        }
        .stat-value {
            font-size: 24px;
            font-weight: 600;
            color: #1a1a1a;
        }
        .controls {
            display: flex;
            gap: 12px;
            margin-bottom: 20px;
            flex-wrap: wrap;
        }
        input, select {
            padding: 8px 12px;
            border: 1px solid #ddd;
            border-radius: 4px;
            font-size: 14px;
        }
        input[type="text"] {
            flex: 1;
            min-width: 200px;
        }
        select {
            min-width: 150px;
        }
        .refresh-btn {
            padding: 8px 16px;
            background: #007bff;
            color: white;
            border: none;
            border-radius: 4px;
            cursor: pointer;
            font-size: 14px;
        }
        .refresh-btn:hover {
            background: #0056b3;
        }
        .refresh-btn:disabled {
            background: #ccc;
            cursor: not-allowed;
        }
        .status-indicator {
            display: inline-block;
            width: 8px;
            height: 8px;
            border-radius: 50%;
            flex-shrink: 0;
        }
        .status-pending { background: #ffc107; }
        .status-running { background: #17a2b8; }
        .status-failed { background: #dc3545; }
        .status-succeeded { background: #28a745; }
        table {
            width: 100%;
            border-collapse: collapse;
            margin-top: 16px;
        }
        th {
            background: #f8f9fa;
            padding: 12px;
            text-align: left;
            font-weight: 600;
            font-size: 12px;
            text-transform: uppercase;
            color: #666;
            border-bottom: 2px solid #dee2e6;
        }
        td {
            padding: 12px;
            border-bottom: 1px solid #dee2e6;
            font-size: 14px;
        }
        .status-cell {
            display: inline-flex;
            align-items: center;
            white-space: nowrap;
            gap: 6px;
            vertical-align: middle;
        }
        tr:hover {
            background: #f8f9fa;
        }
        .empty-state {
            text-align: center;
            padding: 48px;
            color: #999;
        }
        .loading {
            text-align: center;
            padding: 48px;
            color: #666;
        }
        .error {
            background: #f8d7da;
            color: #721c24;
            padding: 12px;
            border-radius: 4px;
            margin-bottom: 16px;
        }
        .badge {
            display: inline-block;
            padding: 4px 8px;
            border-radius: 4px;
            font-size: 12px;
            font-weight: 500;
        }
        .badge-deployment { background: #e7f3ff; color: #0066cc; }
        .badge-statefulset { background: #fff4e6; color: #cc6600; }
        .badge-error { background: #f8d7da; color: #721c24; }
        .badge-warning { background: #fff3cd; color: #856404; }
        .expandable-row {
            cursor: pointer;
        }
        .expandable-row:hover {
            background: #f0f0f0;
        }
        .details-row {
            display: none;
        }
        .details-row.expanded {
            display: table-row;
        }
        .details-content {
            padding: 16px;
            background: #f8f9fa;
            border-left: 4px solid #007bff;
        }
        .details-section {
            margin-bottom: 16px;
        }
        .details-section h4 {
            margin-bottom: 8px;
            color: #333;
            font-size: 14px;
            font-weight: 600;
        }
        .container-error {
            background: white;
            padding: 12px;
            margin-bottom: 8px;
            border-radius: 4px;
            border-left: 3px solid #dc3545;
        }
        .container-error-header {
            font-weight: 600;
            margin-bottom: 4px;
            color: #333;
        }
        .container-error-detail {
            font-size: 12px;
            color: #666;
            margin: 2px 0;
        }
        .pod-condition {
            display: inline-block;
            padding: 4px 8px;
            margin: 2px;
            border-radius: 4px;
            font-size: 12px;
        }
        .condition-true { background: #d4edda; color: #155724; }
        .condition-false { background: #f8d7da; color: #721c24; }
        .condition-unknown { background: #e2e3e5; color: #383d41; }
        .expand-icon {
            display: inline-block;
            width: 12px;
            text-align: center;
            margin-right: 8px;
        }
        .last-update {
            text-align: right;
            color: #999;
            font-size: 12px;
            margin-top: 16px;
        }
    </style>
</head>
<body>
    <div class="container">
        <h1>KubeSleuth Dashboard</h1>
        <div class="subtitle">Monitor non-ready pods across your cluster</div>
        
        <div class="stats">
            <div class="stat-card">
                <div class="stat-label">Total Non-Ready Pods</div>
                <div class="stat-value" id="totalPods">-</div>
            </div>
            <div class="stat-card">
                <div class="stat-label">Namespaces</div>
                <div class="stat-value" id="totalNamespaces">-</div>
            </div>
            <div class="stat-card">
                <div class="stat-label">Deployments Affected</div>
                <div class="stat-value" id="totalDeployments">-</div>
            </div>
        </div>

        <div id="error" class="error" style="display: none;"></div>

        <div class="controls">
            <input type="text" id="search" placeholder="Search pods, namespaces, owners..." oninput="filterTable()">
            <select id="namespaceFilter" onchange="filterTable()">
                <option value="">All Namespaces</option>
            </select>
            <select id="phaseFilter" onchange="filterTable()">
                <option value="">All Phases</option>
                <option value="Pending">Pending</option>
                <option value="Running">Running</option>
                <option value="Failed">Failed</option>
                <option value="Succeeded">Succeeded</option>
            </select>
            <button class="refresh-btn" onclick="loadData()" id="refreshBtn">Refresh</button>
        </div>

        <div id="loading" class="loading">Loading...</div>
        <div id="tableContainer" style="display: none;">
            <table id="podsTable">
                <thead>
                    <tr>
                        <th style="width: 30px;"></th>
                        <th>Pod Name</th>
                        <th>Namespace</th>
                        <th>Phase</th>
                        <th>Owner</th>
                        <th>Reason</th>
                        <th>Message</th>
                    </tr>
                </thead>
                <tbody id="podsTableBody">
                </tbody>
            </table>
        </div>
        <div id="emptyState" class="empty-state" style="display: none;">
            <p>No non-ready pods found. All pods are healthy! 🎉</p>
        </div>
        <div class="last-update" id="lastUpdate"></div>
    </div>

    <script>
        let allPods = [];
        let filteredPods = [];

        async function loadData() {
            const refreshBtn = document.getElementById('refreshBtn');
            const loading = document.getElementById('loading');
            const errorDiv = document.getElementById('error');
            const tableContainer = document.getElementById('tableContainer');
            const emptyState = document.getElementById('emptyState');

            refreshBtn.disabled = true;
            loading.style.display = 'block';
            errorDiv.style.display = 'none';
            tableContainer.style.display = 'none';
            emptyState.style.display = 'none';

            try {
                const response = await fetch('/api/podsleuths');
                if (!response.ok) {
                    throw new Error('Failed to fetch data');
                }
                const data = await response.json();
                
                // Aggregate all non-ready pods from all PodSleuth resources
                allPods = [];
                if (data.items && Array.isArray(data.items) && data.items.length > 0) {
                    data.items.forEach(podSleuth => {
                        if (podSleuth.status && podSleuth.status.nonReadyPods && Array.isArray(podSleuth.status.nonReadyPods)) {
                            allPods = allPods.concat(podSleuth.status.nonReadyPods);
                        }
                    });
                } else if (Array.isArray(data)) {
                    // Fallback: if API returns array directly
                    allPods = data;
                }

                updateStats();
                updateNamespaceFilter();
                filterTable();
                updateLastUpdate();

                loading.style.display = 'none';
                if (filteredPods.length === 0) {
                    emptyState.style.display = 'block';
                } else {
                    tableContainer.style.display = 'block';
                }
            } catch (error) {
                loading.style.display = 'none';
                errorDiv.style.display = 'block';
                errorDiv.textContent = 'Error loading data: ' + error.message;
            } finally {
                refreshBtn.disabled = false;
            }
        }

        function updateStats() {
            const namespaces = new Set(allPods.map(p => p.namespace));
            const deployments = new Set(allPods.filter(p => p.ownerKind === 'Deployment').map(p => p.ownerName));
            
            document.getElementById('totalPods').textContent = allPods.length;
            document.getElementById('totalNamespaces').textContent = namespaces.size;
            document.getElementById('totalDeployments').textContent = deployments.size;
        }

        function updateNamespaceFilter() {
            const namespaces = [...new Set(allPods.map(p => p.namespace))].sort();
            const select = document.getElementById('namespaceFilter');
            const currentValue = select.value;
            
            // Clear and rebuild options
            select.innerHTML = '<option value="">All Namespaces</option>';
            namespaces.forEach(ns => {
                const option = document.createElement('option');
                option.value = ns;
                option.textContent = ns;
                select.appendChild(option);
            });
            
            if (currentValue && namespaces.includes(currentValue)) {
                select.value = currentValue;
            }
        }

        function filterTable() {
            const searchTerm = document.getElementById('search').value.toLowerCase();
            const namespaceFilter = document.getElementById('namespaceFilter').value;
            const phaseFilter = document.getElementById('phaseFilter').value;

            filteredPods = allPods.filter(pod => {
                const matchesSearch = !searchTerm || 
                    pod.name.toLowerCase().includes(searchTerm) ||
                    pod.namespace.toLowerCase().includes(searchTerm) ||
                    (pod.ownerName && pod.ownerName.toLowerCase().includes(searchTerm));
                
                const matchesNamespace = !namespaceFilter || pod.namespace === namespaceFilter;
                const matchesPhase = !phaseFilter || pod.phase === phaseFilter;

                return matchesSearch && matchesNamespace && matchesPhase;
            });

            renderTable();
        }

        function renderTable() {
            const tbody = document.getElementById('podsTableBody');
            tbody.innerHTML = '';

            filteredPods.forEach((pod, index) => {
                const hasDetails = (pod.containerErrors && pod.containerErrors.length > 0) || 
                                  (pod.podConditions && pod.podConditions.length > 0) ||
                                  (pod.logAnalysis && pod.logAnalysis.rootCause);
                
                // Always show expand icon if log analysis is present (it's important)
                const hasLogAnalysis = pod.logAnalysis && pod.logAnalysis.rootCause;
                
                // Main row - make expandable if has details or log analysis
                const row = tbody.insertRow();
                const isExpandable = hasDetails || hasLogAnalysis;
                row.className = isExpandable ? 'expandable-row' : '';
                row.onclick = isExpandable ? () => toggleDetails(index) : null;
                
                // Expand icon - always show if log analysis is present
                const expandCell = row.insertCell(0);
                if (hasDetails || hasLogAnalysis) {
                    const icon = document.createElement('span');
                    icon.className = 'expand-icon';
                    icon.textContent = '▶';
                    icon.id = 'expand-icon-' + index;
                    expandCell.appendChild(icon);
                } else {
                    expandCell.textContent = '';
                }
                
                row.insertCell(1).textContent = pod.name;
                row.insertCell(2).textContent = pod.namespace;
                
                const phaseCell = row.insertCell(3);
                const statusContainer = document.createElement('span');
                statusContainer.className = 'status-cell';
                const statusIndicator = document.createElement('span');
                statusIndicator.className = 'status-indicator status-' + pod.phase.toLowerCase();
                const phaseText = document.createTextNode(pod.phase);
                statusContainer.appendChild(statusIndicator);
                statusContainer.appendChild(phaseText);
                phaseCell.appendChild(statusContainer);
                
                const ownerCell = row.insertCell(4);
                if (pod.ownerKind && pod.ownerName) {
                    const badge = document.createElement('span');
                    badge.className = 'badge badge-' + pod.ownerKind.toLowerCase();
                    badge.textContent = pod.ownerKind + ': ' + pod.ownerName;
                    ownerCell.appendChild(badge);
                } else {
                    ownerCell.textContent = '-';
                }
                
                const reasonCell = row.insertCell(5);
                if (pod.reason) {
                    const badge = document.createElement('span');
                    badge.className = 'badge badge-error';
                    badge.textContent = pod.reason;
                    reasonCell.appendChild(badge);
                } else {
                    reasonCell.textContent = '-';
                }
                
                const messageCell = row.insertCell(6);
                messageCell.style.cssText = 'vertical-align: top; padding: 8px;';
                
                // Extract and highlight log analysis message if present
                let displayMessage = pod.message || '-';
                let logAnalysisMessage = '';
                
                // Check for log analysis in multiple ways (handle both camelCase and PascalCase)
                if (pod.logAnalysis) {
                    logAnalysisMessage = pod.logAnalysis.rootCause || pod.logAnalysis.RootCause || '';
                }
                
                // Extract log analysis from message if it was appended by controller
                // The controller appends ". Log analysis: ..." to the message
                // We want to show both separately: log analysis prominently, then original Kubernetes message
                let originalKubernetesMessage = displayMessage;
                if (displayMessage && typeof displayMessage === 'string' && displayMessage.includes('Log analysis:')) {
                    const parts = displayMessage.split('Log analysis:');
                    if (parts.length > 1) {
                        // If we don't have logAnalysis from object, use the one from message
                        if (!logAnalysisMessage || logAnalysisMessage === '') {
                            logAnalysisMessage = parts[1].trim();
                        }
                        // Get the original Kubernetes message (before log analysis was appended)
                        originalKubernetesMessage = parts[0].trim();
                        // Remove trailing period and space if present
                        if (originalKubernetesMessage.endsWith('.')) {
                            originalKubernetesMessage = originalKubernetesMessage.slice(0, -1).trim();
                        }
                    }
                }
                
                // Build message cell - show original Kubernetes message first, then log analysis
                messageCell.innerHTML = '';
                
                // First line: Original Kubernetes status message (always show if exists)
                if (originalKubernetesMessage && originalKubernetesMessage !== '-' && originalKubernetesMessage !== null && originalKubernetesMessage !== '') {
                    const msgLine = document.createElement('div');
                    msgLine.style.cssText = 'font-size: 12px; color: #666; line-height: 1.4; margin-bottom: 4px;';
                    let msgText = originalKubernetesMessage;
                    if (msgText.length > 100) {
                        msgText = msgText.substring(0, 100) + '...';
                    }
                    msgLine.textContent = msgText;
                    messageCell.appendChild(msgLine);
                } else if (!logAnalysisMessage || logAnalysisMessage === '') {
                    // No log analysis - show message or default
                    if (displayMessage && displayMessage !== '-') {
                        messageCell.textContent = displayMessage.length > 100 ? displayMessage.substring(0, 100) + '...' : displayMessage;
                    } else {
                        messageCell.textContent = '-';
                        messageCell.style.cssText = '';
                    }
                }
                
                // Second line: Log analysis (if present) - at the bottom, small icon
                if (logAnalysisMessage && logAnalysisMessage !== '') {
                    const logAnalysisLine = document.createElement('div');
                    logAnalysisLine.style.cssText = 'line-height: 1.4; padding-top: 4px;';
                    logAnalysisLine.innerHTML = '<span style="color: #ff9800; font-size: 11px;">🔍</span> <span style="font-size: 12px; font-weight: 600; color: #ff9800;">Log analysis:</span> <span style="font-size: 12px; color: #333; font-weight: 500;">' + escapeHtml(logAnalysisMessage) + '</span>';
                    messageCell.appendChild(logAnalysisLine);
                }
                
                // Details row - show if has details or log analysis
                if (hasDetails || hasLogAnalysis) {
                    const detailsRow = tbody.insertRow();
                    detailsRow.className = 'details-row';
                    detailsRow.id = 'details-' + index;
                    const detailsCell = detailsRow.insertCell(0);
                    detailsCell.colSpan = 7;
                    detailsCell.innerHTML = renderDetails(pod);
                }
            });
        }

        function toggleDetails(index) {
            const detailsRow = document.getElementById('details-' + index);
            const icon = document.getElementById('expand-icon-' + index);
            
            if (detailsRow.classList.contains('expanded')) {
                detailsRow.classList.remove('expanded');
                icon.textContent = '▶';
            } else {
                detailsRow.classList.add('expanded');
                icon.textContent = '▼';
            }
        }

        function renderDetails(pod) {
            let html = '<div class="details-content">';
            
            // Container Errors
            if (pod.containerErrors && pod.containerErrors.length > 0) {
                html += '<div class="details-section">';
                html += '<h4>Container Errors (' + pod.containerErrors.length + ')</h4>';
                pod.containerErrors.forEach(err => {
                    html += '<div class="container-error">';
                    html += '<div class="container-error-header">';
                    html += err.containerName + ' (' + err.type + ')';
                    if (err.state) {
                        html += ' - State: ' + err.state;
                    }
                    html += '</div>';
                    if (err.reason) {
                        html += '<div class="container-error-detail"><strong>Reason:</strong> ' + err.reason + '</div>';
                    }
                    if (err.message) {
                        html += '<div class="container-error-detail"><strong>Message:</strong> ' + err.message + '</div>';
                    }
                    if (err.exitCode !== null && err.exitCode !== undefined) {
                        html += '<div class="container-error-detail"><strong>Exit Code:</strong> ' + err.exitCode + '</div>';
                    }
                    if (err.restartCount !== null && err.restartCount !== undefined) {
                        html += '<div class="container-error-detail"><strong>Restart Count:</strong> ' + err.restartCount + '</div>';
                    }
                    html += '<div class="container-error-detail"><strong>Ready:</strong> ' + (err.ready ? 'Yes' : 'No') + '</div>';
                    html += '</div>';
                });
                html += '</div>';
            }
            
            // Pod Conditions
            if (pod.podConditions && pod.podConditions.length > 0) {
                html += '<div class="details-section">';
                html += '<h4>Pod Conditions</h4>';
                pod.podConditions.forEach(condition => {
                    const statusClass = 'condition-' + condition.status.toLowerCase();
                    html += '<span class="pod-condition ' + statusClass + '">';
                    html += condition.type + ': ' + condition.status;
                    if (condition.reason) {
                        html += ' (' + condition.reason + ')';
                    }
                    html += '</span>';
                });
                html += '</div>';
            }
            
            // Log Analysis - Make it prominent and visible
            if (pod.logAnalysis && pod.logAnalysis.rootCause) {
                html += '<div class="details-section" style="border-top: 3px solid #ffc107; padding-top: 16px; margin-top: 16px;">';
                html += '<h4 style="color: #856404; font-size: 16px; margin-bottom: 12px;">🔍 Log Analysis - Root Cause Identified</h4>';
                html += '<div class="container-error" style="background: #fff3cd; border-left: 4px solid #ffc107; padding: 12px;">';
                html += '<div class="container-error-detail" style="font-size: 15px; color: #856404; font-weight: 700; margin-bottom: 8px;">' + escapeHtml(pod.logAnalysis.rootCause) + '</div>';
                
                if (pod.logAnalysis.confidence !== null && pod.logAnalysis.confidence !== undefined) {
                    html += '<div class="container-error-detail"><strong>Confidence:</strong> ' + pod.logAnalysis.confidence + '%</div>';
                }
                
                if (pod.logAnalysis.method) {
                    html += '<div class="container-error-detail"><strong>Method:</strong> ' + pod.logAnalysis.method + '</div>';
                }
                
                if (pod.logAnalysis.analyzedAt) {
                    const analyzedDate = new Date(pod.logAnalysis.analyzedAt);
                    html += '<div class="container-error-detail"><strong>Analyzed At:</strong> ' + analyzedDate.toLocaleString() + '</div>';
                }
                
                if (pod.logAnalysis.errorLines && pod.logAnalysis.errorLines.length > 0) {
                    html += '<div class="container-error-detail" style="margin-top: 8px;"><strong>Error Lines (' + pod.logAnalysis.errorLines.length + '):</strong></div>';
                    html += '<div style="background: #f8f9fa; padding: 8px; border-radius: 4px; margin-top: 4px; max-height: 200px; overflow-y: auto; font-family: monospace; font-size: 11px;">';
                    pod.logAnalysis.errorLines.slice(0, 20).forEach(line => {
                        html += '<div style="margin: 2px 0; color: #721c24;">' + escapeHtml(line) + '</div>';
                    });
                    if (pod.logAnalysis.errorLines.length > 20) {
                        html += '<div style="color: #666; font-style: italic;">... and ' + (pod.logAnalysis.errorLines.length - 20) + ' more lines</div>';
                    }
                    html += '</div>';
                }
                
                html += '</div>';
                html += '</div>';
            }
            
            html += '</div>';
            return html;
        }
        
        function escapeHtml(text) {
            const div = document.createElement('div');
            div.textContent = text;
            return div.innerHTML;
        }

        function updateLastUpdate() {
            const now = new Date();
            document.getElementById('lastUpdate').textContent = 
                'Last updated: ' + now.toLocaleTimeString();
        }

        // Load data on page load
        loadData();
        
        // Auto-refresh every 10 seconds
        setInterval(loadData, 10000);
    </script>
</body>
</html>
`
