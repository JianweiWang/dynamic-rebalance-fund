// 全局变量
let currentBuckets = [];

// 初始化应用
document.addEventListener('DOMContentLoaded', function() {
    loadBuckets();
});

// API调用函数
async function apiCall(url, method = 'GET', data = null) {
    const options = {
        method: method,
        headers: {
            'Content-Type': 'application/json',
        },
    };

    if (data) {
        options.body = JSON.stringify(data);
    }

    try {
        showLoading(true);
        console.log(`API调用: ${method} ${url}`, data); // 调试日志
        
        const response = await fetch(url, options);
        console.log('响应状态:', response.status, response.statusText); // 调试日志
        
        // 检查响应状态
        if (!response.ok) {
            throw new Error(`HTTP ${response.status}: ${response.statusText}`);
        }
        
        // 获取响应文本
        const responseText = await response.text();
        console.log('响应文本:', responseText); // 调试日志
        
        // 检查是否为空响应
        if (!responseText) {
            throw new Error('服务器返回了空响应');
        }
        
        // 尝试解析JSON
        let result;
        try {
            result = JSON.parse(responseText);
        } catch (jsonError) {
            console.error('JSON解析失败:', jsonError, '原始文本:', responseText);
            throw new Error(`JSON解析失败: ${jsonError.message}`);
        }
        
        console.log('解析结果:', result); // 调试日志
        
        if (!result.success) {
            throw new Error(result.message || '操作失败');
        }
        
        return result;
    } catch (error) {
        console.error('API调用失败:', error);
        showMessage('错误: ' + error.message, 'error');
        throw error;
    } finally {
        showLoading(false);
    }
}

// 加载基金数据
async function loadBuckets() {
    try {
        const result = await apiCall('/api/buckets');
        currentBuckets = result.data;
        renderBuckets();
        updateTotalValue();
        populateBucketSelect();
    } catch (error) {
        console.error('加载数据失败:', error);
    }
}

// 渲染基金配置
function renderBuckets() {
    const container = document.getElementById('bucketsContainer');
    container.innerHTML = '';

    currentBuckets.forEach((bucket, bucketIndex) => {
        const bucketDiv = document.createElement('div');
        bucketDiv.className = 'bucket-container fade-in';
        
        const bucketClass = getBucketClass(bucketIndex);
        
        bucketDiv.innerHTML = `
            <div class="bucket-header ${bucketClass}">
                <div>
                    <h6 class="mb-1">${bucket.name}</h6>
                    <small>目标占比: ${(bucket.target_rate * 100).toFixed(1)}%</small>
                </div>
                <div class="text-end">
                    <div class="badge bg-light text-dark">
                        ${bucket.funds.length} 个基金
                    </div>
                </div>
            </div>
            <div class="bucket-content">
                ${bucket.funds.map((fund, fundIndex) => renderFund(fund, bucketIndex, fundIndex)).join('')}
            </div>
        `;
        
        container.appendChild(bucketDiv);
    });
}

// 渲染单个基金
function renderFund(fund, bucketIndex, fundIndex) {
    return `
        <div class="fund-item">
            <div class="fund-header">
                <div class="d-flex align-items-center">
                    <h6 class="fund-name">${fund.name}</h6>
                    <span class="fund-code">${fund.code}</span>
                </div>
                <div class="fund-actions">
                    <button class="btn btn-outline-primary action-btn" 
                            onclick="editFund(${bucketIndex}, ${fundIndex})">
                        <i class="fas fa-edit"></i>
                    </button>
                    <button class="btn btn-outline-danger action-btn" 
                            onclick="deleteFund(${bucketIndex}, ${fundIndex})">
                        <i class="fas fa-trash"></i>
                    </button>
                </div>
            </div>
            <div class="fund-metrics">
                <div class="metric">
                    <div class="metric-label">当前市值</div>
                    <div class="metric-value">${fund.current.toFixed(2)}万</div>
                </div>
                <div class="metric">
                    <div class="metric-label">权重</div>
                    <div class="metric-value">${(fund.weight * 100).toFixed(1)}%</div>
                    <div class="weight-bar">
                        <div class="weight-fill" style="width: ${fund.weight * 100}%"></div>
                    </div>
                </div>
            </div>
        </div>
    `;
}

// 获取桶的样式类
function getBucketClass(index) {
    const classes = ['short-term', 'medium-term', 'long-term'];
    return classes[index] || 'short-term';
}

// 更新总市值
function updateTotalValue() {
    let total = 0;
    currentBuckets.forEach(bucket => {
        bucket.funds.forEach(fund => {
            total += fund.current;
        });
    });
    
    document.getElementById('totalValue').textContent = `总市值: ${total.toFixed(2)}万`;
}

// 填充桶选择器
function populateBucketSelect() {
    const select = document.getElementById('bucketSelect');
    select.innerHTML = '';
    
    currentBuckets.forEach((bucket, index) => {
        const option = document.createElement('option');
        option.value = index;
        option.textContent = bucket.name;
        select.appendChild(option);
    });
}

// 显示添加基金模态框
function showAddFundModal() {
    populateBucketSelect();
    const modal = new bootstrap.Modal(document.getElementById('addFundModal'));
    modal.show();
}

// 添加基金
async function addFund() {
    const bucketIndex = parseInt(document.getElementById('bucketSelect').value);
    const name = document.getElementById('fundName').value.trim();
    const code = document.getElementById('fundCode').value.trim();
    const current = parseFloat(document.getElementById('fundCurrent').value);
    const weight = parseFloat(document.getElementById('fundWeight').value);

    if (!name || !code || isNaN(current) || isNaN(weight)) {
        showMessage('请填写所有必填字段', 'error');
        return;
    }

    if (weight <= 0 || weight > 1) {
        showMessage('权重必须在0-1之间', 'error');
        return;
    }

    try {
        const result = await apiCall('/api/funds', 'POST', {
            bucket_index: bucketIndex,
            name: name,
            code: code,
            current: current,
            weight: weight
        });

        currentBuckets = result.data;
        renderBuckets();
        updateTotalValue();
        
        // 关闭模态框并重置表单
        const modal = bootstrap.Modal.getInstance(document.getElementById('addFundModal'));
        modal.hide();
        document.getElementById('addFundForm').reset();
        
        showMessage('基金添加成功', 'success');
    } catch (error) {
        console.error('添加基金失败:', error);
    }
}

// 编辑基金
function editFund(bucketIndex, fundIndex) {
    const fund = currentBuckets[bucketIndex].funds[fundIndex];
    
    document.getElementById('editBucketIndex').value = bucketIndex;
    document.getElementById('editFundIndex').value = fundIndex;
    document.getElementById('editFundName').value = fund.name;
    document.getElementById('editFundCode').value = fund.code;
    document.getElementById('editFundCurrent').value = fund.current;
    document.getElementById('editFundWeight').value = fund.weight;
    
    const modal = new bootstrap.Modal(document.getElementById('editFundModal'));
    modal.show();
}

// 更新基金
async function updateFund() {
    const bucketIndex = parseInt(document.getElementById('editBucketIndex').value);
    const fundIndex = parseInt(document.getElementById('editFundIndex').value);
    const name = document.getElementById('editFundName').value.trim();
    const code = document.getElementById('editFundCode').value.trim();
    const current = parseFloat(document.getElementById('editFundCurrent').value);
    const weight = parseFloat(document.getElementById('editFundWeight').value);

    if (!name || !code || isNaN(current) || isNaN(weight)) {
        showMessage('请填写所有必填字段', 'error');
        return;
    }

    if (weight <= 0 || weight > 1) {
        showMessage('权重必须在0-1之间', 'error');
        return;
    }

    try {
        // 分别更新每个字段
        const updates = [
            { field: 'name', value: name },
            { field: 'code', value: code },
            { field: 'current', value: current.toString() },
            { field: 'weight', value: weight.toString() }
        ];

        for (const update of updates) {
            const result = await apiCall('/api/funds', 'PUT', {
                bucket_index: bucketIndex,
                fund_index: fundIndex,
                field: update.field,
                value: update.value
            });
            currentBuckets = result.data;
        }

        renderBuckets();
        updateTotalValue();
        
        // 关闭模态框
        const modal = bootstrap.Modal.getInstance(document.getElementById('editFundModal'));
        modal.hide();
        
        showMessage('基金信息更新成功', 'success');
    } catch (error) {
        console.error('更新基金失败:', error);
    }
}

// 删除基金
async function deleteFund(bucketIndex, fundIndex) {
    const fund = currentBuckets[bucketIndex].funds[fundIndex];
    
    if (!confirm(`确定要删除基金 "${fund.name}" 吗？`)) {
        return;
    }

    try {
        const result = await apiCall('/api/funds', 'DELETE', {
            bucket_index: bucketIndex,
            fund_index: fundIndex
        });

        currentBuckets = result.data;
        renderBuckets();
        updateTotalValue();
        
        showMessage('基金删除成功', 'success');
    } catch (error) {
        console.error('删除基金失败:', error);
    }
}

// 显示再平衡模态框
function showRebalanceModal() {
    const threshold = document.getElementById('thresholdInput').value || 0.05;
    performRebalance(parseFloat(threshold));
}

// 执行再平衡
async function performRebalance(threshold = 0.05) {
    try {
        const result = await apiCall('/api/rebalance', 'POST', {
            threshold: threshold
        });

        const rebalanceData = result.data;
        renderRebalanceResults(rebalanceData);
        
        // 显示结果区域
        document.getElementById('rebalanceResultSection').style.display = 'block';
        
        // 滚动到结果区域
        document.getElementById('rebalanceResultSection').scrollIntoView({ 
            behavior: 'smooth' 
        });
        
        showMessage('再平衡分析完成', 'success');
    } catch (error) {
        console.error('再平衡分析失败:', error);
    }
}

// 渲染再平衡结果
function renderRebalanceResults(buckets) {
    const tbody = document.getElementById('rebalanceResults');
    tbody.innerHTML = '';

    buckets.forEach(bucket => {
        bucket.funds.forEach(fund => {
            const row = document.createElement('tr');
            const diffClass = fund.diff > 0 ? 'positive' : fund.diff < 0 ? 'negative' : 'neutral';
            const adviceClass = fund.advice === '买入' ? 'advice-buy' : 
                              fund.advice === '卖出' ? 'advice-sell' : 'advice-hold';
            
            row.innerHTML = `
                <td>
                    <div class="fw-semibold">${fund.name}</div>
                </td>
                <td>
                    <code>${fund.code}</code>
                </td>
                <td>
                    <span class="fw-semibold">${fund.current.toFixed(2)}</span>万
                </td>
                <td>
                    <span class="fw-semibold">${fund.target.toFixed(2)}</span>万
                </td>
                <td>
                    <span class="fw-semibold ${diffClass}">
                        ${fund.diff > 0 ? '+' : ''}${fund.diff.toFixed(2)}万
                    </span>
                </td>
                <td>
                    <span class="fw-semibold ${adviceClass}">
                        ${fund.advice}
                    </span>
                </td>
                <td>
                    <small class="text-muted">${fund.reason || '暂无说明'}</small>
                </td>
            `;
            
            tbody.appendChild(row);
        });
    });
}

// 显示消息
function showMessage(message, type = 'info') {
    const toast = document.getElementById('messageToast');
    const messageElement = document.getElementById('toastMessage');
    
    messageElement.textContent = message;
    
    // 更新图标和颜色
    const icon = toast.querySelector('.toast-header i');
    if (type === 'success') {
        icon.className = 'fas fa-check-circle me-2 text-success';
    } else if (type === 'error') {
        icon.className = 'fas fa-exclamation-circle me-2 text-danger';
    } else {
        icon.className = 'fas fa-info-circle me-2 text-primary';
    }
    
    const bsToast = new bootstrap.Toast(toast);
    bsToast.show();
}

// 显示/隐藏加载指示器
function showLoading(show) {
    const indicator = document.getElementById('loadingIndicator');
    indicator.style.display = show ? 'block' : 'none';
}

// 格式化数字
function formatNumber(num, decimals = 2) {
    return Number(num).toFixed(decimals);
}

// 格式化百分比
function formatPercent(num, decimals = 1) {
    return (num * 100).toFixed(decimals) + '%';
}

// 显示历史记录模态框
async function showHistoryModal() {
    try {
        const result = await apiCall('/api/rebalance/history?limit=20', 'GET');
        console.log('历史记录数据:', result); // 添加调试日志
        renderHistoryRecords(result.data);
        
        const modal = new bootstrap.Modal(document.getElementById('historyModal'));
        modal.show();
    } catch (error) {
        console.error('获取历史记录失败:', error);
        showMessage('获取历史记录失败: ' + error.message, 'error');
    }
}

// 渲染历史记录
function renderHistoryRecords(records) {
    const container = document.getElementById('historyContent');
    
    console.log('渲染历史记录:', records); // 调试日志
    
    if (!records || records.length === 0) {
        container.innerHTML = `
            <div class="text-center py-4">
                <i class="fas fa-inbox text-muted mb-3" style="font-size: 3rem;"></i>
                <p class="text-muted">暂无历史记录</p>
            </div>
        `;
        return;
    }

    let html = `
        <div class="table-responsive">
            <table class="table table-hover">
                <thead class="table-light">
                    <tr>
                        <th>执行时间</th>
                        <th>阈值</th>
                        <th>总市值(万元)</th>
                        <th>操作</th>
                    </tr>
                </thead>
                <tbody>
    `;

    records.forEach(record => {
        const date = new Date(record.created_at).toLocaleString('zh-CN', {
            year: 'numeric',
            month: '2-digit',
            day: '2-digit',
            hour: '2-digit',
            minute: '2-digit',
            second: '2-digit'
        });

        html += `
            <tr>
                <td>
                    <div class="fw-semibold">${date}</div>
                    <small class="text-muted">#${record.id}</small>
                </td>
                <td>
                    <span class="badge bg-primary">${(record.threshold * 100).toFixed(1)}%</span>
                </td>
                <td>
                    <span class="fw-semibold text-success">${record.total_value.toFixed(2)}</span>万
                </td>
                <td>
                    <button class="btn btn-sm btn-outline-info" onclick="viewHistoryDetail(${record.id})">
                        <i class="fas fa-eye me-1"></i>查看详情
                    </button>
                </td>
            </tr>
        `;
    });

    html += `
                </tbody>
            </table>
        </div>
    `;

    container.innerHTML = html;
}

// 查看历史详情
async function viewHistoryDetail(recordId) {
    try {
        const result = await apiCall(`/api/rebalance/history/${recordId}`, 'GET');
        console.log('历史详情数据:', result);
        renderHistoryDetail(result.data);
        
        const modal = new bootstrap.Modal(document.getElementById('historyDetailModal'));
        modal.show();
    } catch (error) {
        console.error('获取历史详情失败:', error);
        showMessage('获取历史详情失败: ' + error.message, 'error');
    }
}

// 渲染历史详情
function renderHistoryDetail(detail) {
    const container = document.getElementById('historyDetailContent');
    const record = detail.record;
    const suggestions = detail.suggestions;

    console.log('渲染历史详情:', detail);

    const date = new Date(record.created_at).toLocaleString('zh-CN', {
        year: 'numeric',
        month: '2-digit',
        day: '2-digit',
        hour: '2-digit',
        minute: '2-digit',
        second: '2-digit'
    });

    // 计算统计信息
    let buyCount = 0, sellCount = 0, holdCount = 0;
    let totalBuyAmount = 0, totalSellAmount = 0;

    suggestions.forEach(suggestion => {
        switch(suggestion.advice) {
            case '买入':
                buyCount++;
                totalBuyAmount += Math.abs(suggestion.diff_value);
                break;
            case '卖出':
                sellCount++;
                totalSellAmount += Math.abs(suggestion.diff_value);
                break;
            case '保持不动':
                holdCount++;
                break;
        }
    });

    let html = `
        <!-- 记录基本信息 -->
        <div class="row mb-4">
            <div class="col-12">
                <div class="card bg-light">
                    <div class="card-body">
                        <div class="row">
                            <div class="col-md-3">
                                <h6 class="card-subtitle mb-1">执行时间</h6>
                                <p class="card-text fw-semibold">${date}</p>
                            </div>
                            <div class="col-md-2">
                                <h6 class="card-subtitle mb-1">阈值</h6>
                                <p class="card-text">
                                    <span class="badge bg-primary">${(record.threshold * 100).toFixed(1)}%</span>
                                </p>
                            </div>
                            <div class="col-md-2">
                                <h6 class="card-subtitle mb-1">总市值</h6>
                                <p class="card-text fw-semibold text-success">${record.total_value.toFixed(2)}万</p>
                            </div>
                            <div class="col-md-2">
                                <h6 class="card-subtitle mb-1">记录ID</h6>
                                <p class="card-text"><code>#${record.id}</code></p>
                            </div>
                        </div>
                    </div>
                </div>
            </div>
        </div>

        <!-- 操作统计 -->
        <div class="row mb-4">
            <div class="col-12">
                <h6 class="mb-3">
                    <i class="fas fa-chart-pie me-2"></i>
                    操作统计
                </h6>
                <div class="row g-3">
                    <div class="col-md-3">
                        <div class="card border-success">
                            <div class="card-body text-center">
                                <i class="fas fa-plus-circle text-success mb-2" style="font-size: 2rem;"></i>
                                <h5 class="card-title text-success">${buyCount}</h5>
                                <p class="card-text">需要买入</p>
                                <small class="text-muted">${totalBuyAmount.toFixed(2)}万</small>
                            </div>
                        </div>
                    </div>
                    <div class="col-md-3">
                        <div class="card border-danger">
                            <div class="card-body text-center">
                                <i class="fas fa-minus-circle text-danger mb-2" style="font-size: 2rem;"></i>
                                <h5 class="card-title text-danger">${sellCount}</h5>
                                <p class="card-text">需要卖出</p>
                                <small class="text-muted">${totalSellAmount.toFixed(2)}万</small>
                            </div>
                        </div>
                    </div>
                    <div class="col-md-3">
                        <div class="card border-secondary">
                            <div class="card-body text-center">
                                <i class="fas fa-pause-circle text-secondary mb-2" style="font-size: 2rem;"></i>
                                <h5 class="card-title text-secondary">${holdCount}</h5>
                                <p class="card-text">保持不动</p>
                            </div>
                        </div>
                    </div>
                    <div class="col-md-3">
                        <div class="card border-info">
                            <div class="card-body text-center">
                                <i class="fas fa-list-alt text-info mb-2" style="font-size: 2rem;"></i>
                                <h5 class="card-title text-info">${suggestions.length}</h5>
                                <p class="card-text">总基金数</p>
                            </div>
                        </div>
                    </div>
                </div>
            </div>
        </div>

        <!-- 详细建议表格 -->
        <div class="row">
            <div class="col-12">
                <h6 class="mb-3">
                    <i class="fas fa-list me-2"></i>
                    详细建议
                </h6>
                <div class="table-responsive">
                    <table class="table table-hover">
                        <thead class="table-light">
                            <tr>
                                <th>基金名称</th>
                                <th>基金代码</th>
                                <th>当前市值(万)</th>
                                <th>目标市值(万)</th>
                                <th>调整金额(万)</th>
                                <th>操作建议</th>
                                <th>详细原因</th>
                            </tr>
                        </thead>
                        <tbody>
    `;

    suggestions.forEach(suggestion => {
        const diffClass = suggestion.diff_value > 0 ? 'positive' : 
                         suggestion.diff_value < 0 ? 'negative' : 'neutral';
        
        const adviceClass = suggestion.advice === '买入' ? 'advice-buy' : 
                           suggestion.advice === '卖出' ? 'advice-sell' : 'advice-hold';

        html += `
            <tr>
                <td>
                    <div class="fw-semibold">${suggestion.fund_name}</div>
                </td>
                <td>
                    <code>${suggestion.fund_code}</code>
                </td>
                <td>
                    <span class="fw-semibold">${suggestion.current_value.toFixed(2)}</span>
                </td>
                <td>
                    <span class="fw-semibold">${suggestion.target_value.toFixed(2)}</span>
                </td>
                <td>
                    <span class="fw-semibold ${diffClass}">
                        ${suggestion.diff_value > 0 ? '+' : ''}${suggestion.diff_value.toFixed(2)}
                    </span>
                </td>
                <td>
                    <span class="fw-semibold ${adviceClass}">
                        ${suggestion.advice}
                    </span>
                </td>
                <td>
                    <small class="text-muted">${suggestion.reason || '暂无说明'}</small>
                </td>
            </tr>
        `;
    });

    html += `
                        </tbody>
                    </table>
                </div>
            </div>
        </div>
    `;

    container.innerHTML = html;
}
