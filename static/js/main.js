document.addEventListener('DOMContentLoaded', function() {
    const urlInputs = document.getElementById('urlInputs');
    const addUrlBtn = document.getElementById('addUrlBtn');
    const validateBtn = document.getElementById('validateBtn');
    const template = document.getElementById('urlInputTemplate');

    // 添加链接输入框
    function addUrlInput() {
        const clone = template.content.cloneNode(true);
        urlInputs.appendChild(clone);
        
        const newInput = urlInputs.lastElementChild;
        const input = newInput.querySelector('input');
        const removeBtn = newInput.querySelector('.btn-remove');

        input.addEventListener('input', validateInput);
        removeBtn.addEventListener('click', () => removeUrlInput(newInput));

        validateAllInputs();
        return newInput;
    }

    // 删除链接输入框
    function removeUrlInput(inputGroup) {
        if (urlInputs.children.length > 1) {
            inputGroup.classList.add('removing');
            setTimeout(() => {
                inputGroup.remove();
                validateAllInputs();
            }, 300);
        }
    }

    // 验证单个输入框
    function validateInput(event) {
        const input = event.target;
        const inputGroup = input.closest('.url-input-group');
        const url = input.value.trim();

        try {
            const urlObj = new URL(url);
            const isValid = (urlObj.protocol === 'http:' || urlObj.protocol === 'https:') 
                && (url.toLowerCase().endsWith('.m3u') || url.toLowerCase().endsWith('.m3u8'));
            
            inputGroup.classList.toggle('is-invalid', !isValid && url !== '');
            validateBtn.disabled = !isValid;
            return isValid;
        } catch {
            inputGroup.classList.toggle('is-invalid', url !== '');
            validateBtn.disabled = true;
            return false;
        }
    }

    // 验证所有输入框
    function validateAllInputs() {
        const inputs = urlInputs.querySelectorAll('input');
        const validUrls = Array.from(inputs)
            .map(input => input.value.trim())
            .filter(url => {
                try {
                    const urlObj = new URL(url);
                    return (urlObj.protocol === 'http:' || urlObj.protocol === 'https:') 
                        && (url.toLowerCase().endsWith('.m3u') || url.toLowerCase().endsWith('.m3u8'));
                } catch {
                    return false;
                }
            });

        validateBtn.disabled = validUrls.length === 0;
        return validUrls;
    }

    // 显示结果
    function showResult(message) {
        const result = document.getElementById('result');
        result.innerHTML = message;
    }

    // 复制URL
    function copyUrl(url) {
        navigator.clipboard.writeText(url).then(() => {
            const btn = document.querySelector('.btn-outline-success');
            const originalHtml = btn.innerHTML;
            
            // 更改图标为对勾
            btn.innerHTML = `
                <svg xmlns="http://www.w3.org/2000/svg" width="16" height="16" fill="currentColor" class="bi bi-check" viewBox="0 0 16 16">
                    <path d="M10.97 4.97a.75.75 0 0 1 1.07 1.05l-3.99 4.99a.75.75 0 0 1-1.08.02L4.324 8.384a.75.75 0 1 1 1.06-1.06l2.094 2.093 3.473-4.425a.267.267 0 0 1 .02-.022z"/>
                </svg>
            `;
            
            // 1秒后恢复原始图标
            setTimeout(() => {
                btn.innerHTML = originalHtml;
            }, 1000);
        }).catch(err => {
            console.error('复制失败:', err);
            alert('复制失败，请手动复制链接');
        });
    }

    // 更新进度文本
    function updateProgressText(progress, totalUrls) {
        const progressText = document.getElementById('progressText');
        if (progress < 30) {
            progressText.textContent = `正在获取 ${totalUrls} 个 M3U 文件...`;
        } else if (progress < 60) {
            progressText.textContent = '正在验证链接...';
        } else if (progress < 90) {
            progressText.textContent = '正在合并并生成缓存文件...';
        }
    }

    // 验证并提交
    function validateM3U() {
        const validUrls = validateAllInputs();
        if (validUrls.length === 0) return;

        const validateBtn = document.getElementById('validateBtn');
        const validateBtnText = document.getElementById('validateBtnText');
        const validateSpinner = document.getElementById('validateSpinner');
        const progressArea = document.getElementById('progressArea');
        const progressBar = document.getElementById('progressBar');
        const progressText = document.getElementById('progressText');

        // 清空之前的结果
        document.getElementById('result').innerHTML = '';

        // 显示进度区域
        progressArea.classList.remove('d-none');
        progressBar.style.width = '0%';
        progressText.textContent = `正在处理 ${validUrls.length} 个链接...`;
        
        // 禁用按钮并显示加载动画
        validateBtn.disabled = true;
        validateBtnText.textContent = '验证中...';
        validateSpinner.classList.remove('d-none');

        // 模拟进度
        let progress = 0;
        const progressInterval = setInterval(() => {
            if (progress < 90) {
                progress += 5;
                progressBar.style.width = progress + '%';
                updateProgressText(progress, validUrls.length);
            }
        }, 500);

        // 发送请求到后端
        fetch('/api/validate', {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
            },
            body: JSON.stringify({ urls: validUrls })
        })
        .then(response => response.json())
        .then(data => {
            clearInterval(progressInterval);
            progressBar.style.width = '100%';
            progressText.textContent = '处理完成！';
            
            if (data.success) {
                const currentHost = window.location.origin;
                const m3uUrl = `${currentHost}/iptv.m3u`;
                showResult(`
                    <div class="alert alert-success" role="alert">
                        <h4 class="alert-heading">验证成功！</h4>
                        <p>已成功处理 ${validUrls.length} 个链接，您可以通过以下地址访问合并后的 M3U 文件：</p>
                        <hr>
                        <div class="d-flex align-items-center justify-content-center gap-2">
                            <a href="${m3uUrl}" class="alert-link">${m3uUrl}</a>
                            <button class="btn btn-sm btn-outline-success" onclick="copyUrl('${m3uUrl}')" title="复制链接">
                                <svg xmlns="http://www.w3.org/2000/svg" width="16" height="16" fill="currentColor" class="bi bi-clipboard" viewBox="0 0 16 16">
                                    <path d="M4 1.5H3a2 2 0 0 0-2 2V14a2 2 0 0 0 2 2h10a2 2 0 0 0 2-2V3.5a2 2 0 0 0-2-2h-1v1h1a1 1 0 0 1 1 1V14a1 1 0 0 1-1 1H3a1 1 0 0 1-1-1V3.5a1 1 0 0 1 1-1h1v-1z"/>
                                    <path d="M9.5 1a.5.5 0 0 1 .5.5v1a.5.5 0 0 1-.5.5h-3a.5.5 0 0 1-.5-.5v-1a.5.5 0 0 1 .5-.5h3zm-3-1A1.5 1.5 0 0 0 5 1.5v1A1.5 1.5 0 0 0 6.5 4h3A1.5 1.5 0 0 0 11 2.5v-1A1.5 1.5 0 0 0 9.5 0h-3z"/>
                                </svg>
                            </button>
                        </div>
                    </div>
                `);
            } else {
                showResult(`
                    <div class="alert alert-danger" role="alert">
                        <h4 class="alert-heading">处理失败</h4>
                        <p>${data.message || '处理失败，请稍后重试'}</p>
                        <hr>
                        <p class="mb-0">请检查链接是否正确，或稍后重试。</p>
                    </div>
                `);
            }
        })
        .catch(error => {
            clearInterval(progressInterval);
            showResult(`
                <div class="alert alert-danger" role="alert">
                    <h4 class="alert-heading">处理失败</h4>
                    <p>${error.message || '发生未知错误，请稍后重试'}</p>
                    <hr>
                    <p class="mb-0">请检查链接是否正确，或稍后重试。</p>
                </div>
            `);
        })
        .finally(() => {
            validateBtn.disabled = false;
            validateBtnText.textContent = '验证';
            validateSpinner.classList.add('d-none');
            setTimeout(() => {
                progressArea.classList.add('d-none');
            }, 1000);
        });
    }

    // 事件监听
    addUrlBtn.addEventListener('click', addUrlInput);
    validateBtn.addEventListener('click', validateM3U);

    // 初始化第一个输入框，但不显示删除按钮
    const firstInput = addUrlInput();
    firstInput.querySelector('.btn-remove').style.display = 'none';
    
    // 为第一个输入框添加验证事件
    const firstInputField = firstInput.querySelector('input');
    firstInputField.addEventListener('input', validateInput);
});

// ... 保留之前的其他函数 ... 