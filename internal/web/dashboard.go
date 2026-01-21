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
        @keyframes pulse {
            0%, 100% {
                opacity: 1;
            }
            50% {
                opacity: 0.85;
            }
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
        .refresh-status {
            display: inline-block;
            margin-left: 8px;
            padding: 2px 6px;
            border-radius: 3px;
            font-size: 11px;
            background: #fff3cd;
            color: #856404;
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
        <div class="last-update">
            <span id="lastUpdate"></span>
            <span id="refreshStatus" class="refresh-status" style="display: none;">Auto-refresh paused</span>
        </div>
    </div>

    <script>
        let allPods = [];
        let filteredPods = [];
        let expandedRows = new Set(); // Track which rows are expanded
        let autoRefreshIntervalId = null; // Store interval ID for auto-refresh

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
            // Save currently expanded rows before re-rendering
            const currentlyExpanded = new Set(expandedRows);
            
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
                
                // Second line: Log analysis clickable link (if present)
                if (pod.logAnalysis && (pod.logAnalysis.patternResult || pod.logAnalysis.aiResult)) {
                    const logAnalysisLink = document.createElement('div');
                    logAnalysisLink.style.cssText = 'margin-top: 8px; padding: 8px; background: #fff3cd; border-left: 3px solid #ffc107; border-radius: 4px; cursor: pointer; transition: background 0.2s;';
                    logAnalysisLink.onmouseover = function() { this.style.background = '#ffe69c'; };
                    logAnalysisLink.onmouseout = function() { this.style.background = '#fff3cd'; };
                    logAnalysisLink.onclick = function(e) {
                        e.stopPropagation();
                        toggleDetails(index);
                        // Scroll to details after a short delay
                        setTimeout(() => {
                            const detailsRow = document.getElementById('details-' + index);
                            if (detailsRow && detailsRow.classList.contains('expanded')) {
                                detailsRow.scrollIntoView({ behavior: 'smooth', block: 'nearest' });
                            }
                        }, 100);
                    };
                    
                    // Build summary
                    let summaryParts = [];
                    if (pod.logAnalysis.patternResult && pod.logAnalysis.patternResult.rootCause) {
                        summaryParts.push('Pattern: ' + pod.logAnalysis.patternResult.matchedPattern);
                    }
                    if (pod.logAnalysis.aiResult && pod.logAnalysis.aiResult.rootCause) {
                        summaryParts.push('AI: ' + pod.logAnalysis.aiResult.model);
                    }
                    
                    logAnalysisLink.innerHTML = '<div style="display: flex; align-items: center; gap: 8px;">' +
                        '<span style="font-size: 16px;">🔍</span>' +
                        '<div style="flex: 1;">' +
                        '<strong style="color: #856404; font-size: 13px;">Log analysis found something. Click here to view it.</strong>' +
                        (summaryParts.length > 0 ? '<div style="font-size: 11px; color: #856404; margin-top: 2px;">(' + summaryParts.join(' • ') + ')</div>' : '') +
                        '</div>' +
                        '</div>';
                    
                    messageCell.appendChild(logAnalysisLink);
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
            
            // Restore expanded state after re-rendering
            currentlyExpanded.forEach(index => {
                const detailsRow = document.getElementById('details-' + index);
                const icon = document.getElementById('expand-icon-' + index);
                if (detailsRow && icon) {
                    detailsRow.classList.add('expanded');
                    icon.textContent = '▼';
                }
            });
        }

        function toggleDetails(index) {
            const detailsRow = document.getElementById('details-' + index);
            const icon = document.getElementById('expand-icon-' + index);
            
            if (detailsRow.classList.contains('expanded')) {
                // Closing details
                detailsRow.classList.remove('expanded');
                icon.textContent = '▶';
                expandedRows.delete(index);
                
                // If no more expanded rows, resume auto-refresh
                if (expandedRows.size === 0) {
                    resumeAutoRefresh();
                }
            } else {
                // Opening details
                detailsRow.classList.add('expanded');
                icon.textContent = '▼';
                expandedRows.add(index);
                
                // Pause auto-refresh when any row is expanded
                pauseAutoRefresh();
            }
        }

        function pauseAutoRefresh() {
            if (autoRefreshIntervalId !== null) {
                clearInterval(autoRefreshIntervalId);
                autoRefreshIntervalId = null;
                document.getElementById('refreshStatus').style.display = 'inline-block';
            }
        }

        function resumeAutoRefresh() {
            if (autoRefreshIntervalId === null) {
                document.getElementById('refreshStatus').style.display = 'none';
                // Start auto-refresh immediately and then every 10 seconds
                autoRefreshIntervalId = setInterval(loadData, 10000);
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
            
            // Log Analysis - Always Visible in Details
            if (pod.logAnalysis && (pod.logAnalysis.patternResult || pod.logAnalysis.aiResult)) {
                html += '<div class="details-section" style="border-top: 3px solid #ffc107; padding-top: 16px; margin-top: 16px;">';
                html += '<h4 style="color: #856404; font-size: 16px; margin-bottom: 12px;">🔍 Log Analysis Results</h4>';
                
                // Common Log Analysis Information (MOVED TO TOP)
                html += '<div class="details-section" style="background: #f8f9fa; padding: 12px; border-radius: 4px; margin-bottom: 16px;">';
                
                if (pod.logAnalysis.methods && pod.logAnalysis.methods.length > 0) {
                    html += '<div class="container-error-detail" style="margin-bottom: 4px;"><strong>Methods Used:</strong> ' + pod.logAnalysis.methods.join(', ') + '</div>';
                }
                
                if (pod.logAnalysis.analyzedAt) {
                    const analyzedDate = new Date(pod.logAnalysis.analyzedAt);
                    html += '<div class="container-error-detail" style="margin-bottom: 4px;"><strong>Analyzed At:</strong> ' + analyzedDate.toLocaleString() + '</div>';
                }
                
                if (pod.logAnalysis.cachedAt) {
                    const cachedDate = new Date(pod.logAnalysis.cachedAt);
                    html += '<div class="container-error-detail"><strong>Cached At:</strong> ' + cachedDate.toLocaleString() + ' <span style="color: #28a745; font-weight: 600;">✓ Cached</span></div>';
                }
                
                html += '</div>';
                
                // Pattern Analysis
                if (pod.logAnalysis.patternResult) {
                    html += '<div class="details-section" style="border-top: 2px solid #17a2b8; padding-top: 12px; margin-top: 12px;">';
                    html += '<h4 style="color: #0c5460; font-size: 16px; margin-bottom: 12px;">🔍 Pattern Analysis</h4>';
                    
                    if (pod.logAnalysis.patternResult.error) {
                        html += '<div class="container-error" style="background: #f8d7da; border-left: 4px solid #dc3545; padding: 12px;">';
                        html += '<div style="display: flex; align-items: center; gap: 8px; margin-bottom: 8px;">';
                        html += '<span style="font-size: 24px;">⚠️</span>';
                        html += '<strong style="color: #721c24; font-size: 16px;">Pattern Analysis Failed</strong>';
                        html += '</div>';
                        html += '<div class="container-error-detail" style="font-size: 14px; color: #721c24; font-family: monospace; background: #fff; padding: 8px; border-radius: 4px;">' + escapeHtml(pod.logAnalysis.patternResult.error) + '</div>';
                        html += '</div>';
                    } else {
                        html += '<div class="container-error" style="background: #d1ecf1; border-left: 4px solid #17a2b8; padding: 12px;">';
                        
                        if (pod.logAnalysis.patternResult.rootCause) {
                            html += '<div class="container-error-detail" style="font-size: 15px; color: #0c5460; font-weight: 700; margin-bottom: 8px;">' + escapeHtml(pod.logAnalysis.patternResult.rootCause) + '</div>';
                        }
                        
                        if (pod.logAnalysis.patternResult.matchedPattern) {
                            html += '<div class="container-error-detail"><strong>Matched Pattern:</strong> ' + escapeHtml(pod.logAnalysis.patternResult.matchedPattern) + '</div>';
                        }
                        
                        if (pod.logAnalysis.patternResult.confidence !== null && pod.logAnalysis.patternResult.confidence !== undefined) {
                            html += '<div class="container-error-detail"><strong>Confidence:</strong> ' + pod.logAnalysis.patternResult.confidence + '%</div>';
                        }
                        
                        if (pod.logAnalysis.patternResult.priority !== null && pod.logAnalysis.patternResult.priority !== undefined) {
                            html += '<div class="container-error-detail"><strong>Priority:</strong> ' + pod.logAnalysis.patternResult.priority + '</div>';
                        }
                        
                        html += '</div>';
                    }
                    
                    html += '</div>';
                }
                
                // AI Analysis
                if (pod.logAnalysis.aiResult) {
                    html += '<div class="details-section" style="border-top: 2px solid #6f42c1; padding-top: 12px; margin-top: 12px;">';
                    html += '<h4 style="color: #4c2a85; font-size: 16px; margin-bottom: 12px;">🤖 AI Analysis</h4>';
                    
                    if (pod.logAnalysis.aiResult.error) {
                        html += '<div class="container-error" style="background: #f8d7da; border-left: 4px solid #dc3545; padding: 12px; animation: pulse 2s ease-in-out infinite;">';
                        html += '<div style="display: flex; align-items: center; gap: 8px; margin-bottom: 8px;">';
                        html += '<span style="font-size: 24px;">❌</span>';
                        html += '<strong style="color: #721c24; font-size: 16px;">AI Analysis Failed</strong>';
                        html += '</div>';
                        html += '<div class="container-error-detail" style="font-size: 14px; color: #721c24; font-family: monospace; background: #fff; padding: 8px; border-radius: 4px; white-space: pre-wrap;">' + escapeHtml(pod.logAnalysis.aiResult.error) + '</div>';
                        html += '<div style="margin-top: 8px; padding: 8px; background: #fff3cd; border-radius: 4px; font-size: 12px; color: #856404;">';
                        html += '💡 <strong>Tip:</strong> Check your AI configuration (model name, endpoint, API key)';
                        html += '</div>';
                        html += '</div>';
                    } else {
                        html += '<div class="container-error" style="background: #e7e3f4; border-left: 4px solid #6f42c1; padding: 12px;">';
                        
                        if (pod.logAnalysis.aiResult.rootCause) {
                            html += '<div class="container-error-detail" style="font-size: 15px; color: #4c2a85; font-weight: 700; margin-bottom: 8px;">' + escapeHtml(pod.logAnalysis.aiResult.rootCause) + '</div>';
                        }
                        
                        if (pod.logAnalysis.aiResult.model) {
                            html += '<div class="container-error-detail"><strong>Model:</strong> ' + escapeHtml(pod.logAnalysis.aiResult.model) + '</div>';
                        }
                        
                        if (pod.logAnalysis.aiResult.confidence !== null && pod.logAnalysis.aiResult.confidence !== undefined) {
                            html += '<div class="container-error-detail"><strong>Confidence:</strong> ' + pod.logAnalysis.aiResult.confidence + '%</div>';
                        }
                        
                        html += '</div>';
                    }
                    
                    html += '</div>';
                }
                
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
        
        // Start auto-refresh every 10 seconds
        autoRefreshIntervalId = setInterval(loadData, 10000);
    </script>
</body>
</html>
`
