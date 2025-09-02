// 投资组合概览页面 JavaScript
let portfolios = [];

// 初始化应用
document.addEventListener('DOMContentLoaded', function() {
    loadPortfolios();
    
    // 为比例输入框添加实时计算监听器
    const ratioInputs = ['shortTermRatio', 'mediumTermRatio', 'longTermRatio'];
    ratioInputs.forEach(inputId => {
        const input = document.getElementById(inputId);
        if (input) {
            input.addEventListener('input', updateTotalRatio);
        }
    });
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
        console.log(`API调用: ${method} ${url}`, data);
        
        const response = await fetch(url, options);
        console.log('响应状态:', response.status, response.statusText);
        
        if (!response.ok) {
            throw new Error(`HTTP ${response.status}: ${response.statusText}`);
        }
        
        const responseText = await response.text();
        console.log('响应文本:', responseText);
        
        if (!responseText) {
            throw new Error('服务器返回了空响应');
        }
        
        let result;
        try {
            result = JSON.parse(responseText);
        } catch (jsonError) {
            console.error('JSON解析失败:', jsonError, '原始文本:', responseText);
            throw new Error(`JSON解析失败: ${jsonError.message}`);
        }
        
        console.log('解析结果:', result);
        
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

// 加载投资组合列表
async function loadPortfolios() {
    try {
        const result = await apiCall('/api/portfolios');
        portfolios = result.data || [];
        renderPortfolios();
    } catch (error) {
        console.error('加载投资组合失败:', error);
        showEmptyState();
    }
}

// 渲染投资组合列表
function renderPortfolios() {
    const container = document.getElementById('portfoliosContainer');
    const emptyState = document.getElementById('emptyState');
    
    if (!portfolios || portfolios.length === 0) {
        container.innerHTML = '';
        emptyState.style.display = 'block';
        return;
    }
    
    emptyState.style.display = 'none';
    container.innerHTML = '';
    
    portfolios.forEach(portfolio => {
        const portfolioCard = createPortfolioCard(portfolio);
        container.appendChild(portfolioCard);
    });
}

// 更新总占比显示
function updateTotalRatio() {
    const shortTerm = parseFloat(document.getElementById('shortTermRatio').value) || 0;
    const mediumTerm = parseFloat(document.getElementById('mediumTermRatio').value) || 0;
    const longTerm = parseFloat(document.getElementById('longTermRatio').value) || 0;
    
    const total = shortTerm + mediumTerm + longTerm;
    const display = document.getElementById('totalRatioDisplay');
    
    if (display) {
        display.textContent = `${total.toFixed(1)}%`;
        
        // 颜色提示
        if (Math.abs(total - 100) < 0.1) {
            display.className = 'fw-bold text-success';
        } else {
            display.className = 'fw-bold text-danger';
        }
    }
}

// 创建投资组合卡片
function createPortfolioCard(portfolio) {
    const col = document.createElement('div');
    col.className = 'col-md-6 col-lg-4 mb-4';
    
    const createdDate = new Date(portfolio.created_at).toLocaleDateString('zh-CN');
    const updatedDate = new Date(portfolio.updated_at).toLocaleDateString('zh-CN');
    const totalAmount = portfolio.total_investment_amount || 0;
    
    col.innerHTML = `
        <div class="card h-100 shadow-sm portfolio-card">
            <div class="card-header bg-primary text-white">
                <div class="d-flex justify-content-between align-items-center">
                    <h6 class="mb-0">
                        <i class="fas fa-briefcase me-2"></i>
                        ${portfolio.name}
                    </h6>
                    <div class="dropdown">
                        <button class="btn btn-link text-white p-0" type="button" data-bs-toggle="dropdown">
                            <i class="fas fa-ellipsis-v"></i>
                        </button>
                        <ul class="dropdown-menu">
                            <li><a class="dropdown-item" href="#" onclick="editPortfolio(${portfolio.id})">
                                <i class="fas fa-edit me-2"></i>编辑
                            </a></li>
                            <li><hr class="dropdown-divider"></li>
                            <li><a class="dropdown-item text-danger" href="#" onclick="deletePortfolio(${portfolio.id}, '${portfolio.name}')">
                                <i class="fas fa-trash me-2"></i>删除
                            </a></li>
                        </ul>
                    </div>
                </div>
            </div>
            <div class="card-body">
                <p class="card-text text-muted">${portfolio.description || '暂无描述'}</p>
                <div class="portfolio-stats">
                    <div class="row text-center">
                        <div class="col-12 mb-2">
                            <div class="stat-item bg-light">
                                <div class="stat-label">投资金额</div>
                                <div class="stat-value text-primary">¥${(totalAmount * 10000).toLocaleString()}元</div>
                            </div>
                        </div>
                        <div class="col-6">
                            <div class="stat-item">
                                <div class="stat-label">创建时间</div>
                                <div class="stat-value">${createdDate}</div>
                            </div>
                        </div>
                        <div class="col-6">
                            <div class="stat-item">
                                <div class="stat-label">更新时间</div>
                                <div class="stat-value">${updatedDate}</div>
                            </div>
                        </div>
                    </div>
                </div>
            </div>
            <div class="card-footer bg-transparent">
                <div class="d-grid">
                    <a href="/portfolios/${portfolio.id}" class="btn btn-primary">
                        <i class="fas fa-eye me-2"></i>
                        查看详情
                    </a>
                </div>
            </div>
        </div>
    `;
    
    return col;
}

// 显示空状态
function showEmptyState() {
    const container = document.getElementById('portfoliosContainer');
    const emptyState = document.getElementById('emptyState');
    
    container.innerHTML = '';
    emptyState.style.display = 'block';
}

// 显示创建投资组合模态框
function showCreatePortfolioModal() {
    document.getElementById('portfolioName').value = '';
    document.getElementById('portfolioDescription').value = '';
    document.getElementById('portfolioTotalAmount').value = '';
    document.getElementById('shortTermRatio').value = '10';
    document.getElementById('mediumTermRatio').value = '30';
    document.getElementById('longTermRatio').value = '60';
    
    // 更新总占比显示
    updateTotalRatio();
    
    const modal = new bootstrap.Modal(document.getElementById('createPortfolioModal'));
    modal.show();
}

// 创建投资组合
async function createPortfolio() {
    const name = document.getElementById('portfolioName').value.trim();
    const description = document.getElementById('portfolioDescription').value.trim();
    const totalAmount = parseFloat(document.getElementById('portfolioTotalAmount').value);
    const shortTermRatio = parseFloat(document.getElementById('shortTermRatio').value) / 100;
    const mediumTermRatio = parseFloat(document.getElementById('mediumTermRatio').value) / 100;
    const longTermRatio = parseFloat(document.getElementById('longTermRatio').value) / 100;
    
    // 基本验证
    if (!name) {
        showMessage('请输入投资组合名称', 'error');
        return;
    }
    
    if (isNaN(totalAmount) || totalAmount <= 0) {
        showMessage('请输入有效的初始投资金额', 'error');
        return;
    }
    
    // 验证比例
    if (isNaN(shortTermRatio) || isNaN(mediumTermRatio) || isNaN(longTermRatio)) {
        showMessage('请输入有效的投资占比', 'error');
        return;
    }
    
    if (shortTermRatio < 0 || mediumTermRatio < 0 || longTermRatio < 0) {
        showMessage('投资占比不能为负数', 'error');
        return;
    }
    
    const totalRatio = shortTermRatio + mediumTermRatio + longTermRatio;
    if (Math.abs(totalRatio - 1.0) > 0.001) {
        showMessage('三桶投资占比总和必须等于100%，当前为' + (totalRatio * 100).toFixed(1) + '%', 'error');
        return;
    }
    
    try {
        const result = await apiCall('/api/portfolios', 'POST', {
            name: name,
            description: description,
            total_investment_amount: totalAmount / 10000, // 转换为万元存储
            short_term_ratio: shortTermRatio,
            medium_term_ratio: mediumTermRatio,
            long_term_ratio: longTermRatio
        });
        
        showMessage(result.message, 'success');
        
        // 关闭模态框
        const modal = bootstrap.Modal.getInstance(document.getElementById('createPortfolioModal'));
        modal.hide();
        
        // 刷新列表
        await loadPortfolios();
        
    } catch (error) {
        console.error('创建投资组合失败:', error);
    }
}

// 编辑投资组合
async function editPortfolio(portfolioId) {
    const portfolio = portfolios.find(p => p.id === portfolioId);
    if (!portfolio) {
        showMessage('未找到投资组合', 'error');
        return;
    }
    
    document.getElementById('editPortfolioId').value = portfolio.id;
    document.getElementById('editPortfolioName').value = portfolio.name;
    document.getElementById('editPortfolioDescription').value = portfolio.description || '';
    document.getElementById('editPortfolioTotalAmount').value = (portfolio.total_investment_amount || 0) * 10000; // 转换为元显示
    
    const modal = new bootstrap.Modal(document.getElementById('editPortfolioModal'));
    modal.show();
}

// 更新投资组合
async function updatePortfolio() {
    const id = parseInt(document.getElementById('editPortfolioId').value);
    const name = document.getElementById('editPortfolioName').value.trim();
    const description = document.getElementById('editPortfolioDescription').value.trim();
    const totalAmount = parseFloat(document.getElementById('editPortfolioTotalAmount').value) || 0;
    
    if (!name) {
        showMessage('请输入投资组合名称', 'error');
        return;
    }
    
    if (totalAmount < 0) {
        showMessage('投资金额不能为负数', 'error');
        return;
    }
    
    try {
        const result = await apiCall('/api/portfolios', 'PUT', {
            id: id,
            name: name,
            description: description,
            total_investment_amount: totalAmount / 10000 // 转换为万元存储
        });
        
        showMessage(result.message, 'success');
        
        // 关闭模态框
        const modal = bootstrap.Modal.getInstance(document.getElementById('editPortfolioModal'));
        modal.hide();
        
        // 刷新列表
        await loadPortfolios();
        
    } catch (error) {
        console.error('更新投资组合失败:', error);
    }
}

// 删除投资组合
function deletePortfolio(portfolioId, portfolioName) {
    document.getElementById('deletePortfolioId').value = portfolioId;
    document.getElementById('deletePortfolioName').textContent = portfolioName;
    
    const modal = new bootstrap.Modal(document.getElementById('deletePortfolioModal'));
    modal.show();
}

// 确认删除投资组合
async function confirmDeletePortfolio() {
    const portfolioId = parseInt(document.getElementById('deletePortfolioId').value);
    
    try {
        const result = await apiCall(`/api/portfolios/${portfolioId}`, 'DELETE');
        
        showMessage(result.message, 'success');
        
        // 关闭模态框
        const modal = bootstrap.Modal.getInstance(document.getElementById('deletePortfolioModal'));
        modal.hide();
        
        // 刷新列表
        await loadPortfolios();
        
    } catch (error) {
        console.error('删除投资组合失败:', error);
    }
}

// 显示加载状态
function showLoading(show) {
    const indicator = document.getElementById('loadingIndicator');
    if (indicator) {
        indicator.style.display = show ? 'block' : 'none';
    }
}

// 显示消息
function showMessage(message, type = 'info') {
    const toast = document.getElementById('messageToast');
    const toastMessage = document.getElementById('toastMessage');
    const toastHeader = toast.querySelector('.toast-header i');
    
    // 设置消息内容
    toastMessage.textContent = message;
    
    // 设置图标和样式
    toastHeader.className = 'me-2';
    switch (type) {
        case 'success':
            toastHeader.classList.add('fas', 'fa-check-circle', 'text-success');
            break;
        case 'error':
            toastHeader.classList.add('fas', 'fa-exclamation-circle', 'text-danger');
            break;
        case 'warning':
            toastHeader.classList.add('fas', 'fa-exclamation-triangle', 'text-warning');
            break;
        default:
            toastHeader.classList.add('fas', 'fa-info-circle', 'text-primary');
    }
    
    // 显示toast
    const bsToast = new bootstrap.Toast(toast);
    bsToast.show();
}