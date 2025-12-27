package auth

const setupTemplate = `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Airwallex CLI - Connect Account</title>
    <link rel="preconnect" href="https://fonts.googleapis.com">
    <link rel="preconnect" href="https://fonts.gstatic.com" crossorigin>
    <link href="https://fonts.googleapis.com/css2?family=Inter:wght@400;500;600;700&family=JetBrains+Mono:wght@400;500&display=swap" rel="stylesheet">
    <style>
        :root {
            --bg: #F8F9FB;
            --bg-card: #FFFFFF;
            --bg-hint: #F3F4F8;
            --bg-input: #F5F6FA;
            --border: #E5E7EB;
            --border-focus: #7C3AED;
            --text: #1F2937;
            --text-secondary: #6B7280;
            --text-muted: #9CA3AF;
            --primary: #7C3AED;
            --primary-light: #EDE9FE;
            --coral: #FF6B54;
            --coral-light: #FFF1EF;
            --success: #10B981;
            --success-light: #D1FAE5;
            --error: #EF4444;
            --error-light: #FEE2E2;
        }

        * {
            margin: 0;
            padding: 0;
            box-sizing: border-box;
        }

        html {
            height: 100%;
        }

        body {
            font-family: 'Inter', -apple-system, BlinkMacSystemFont, sans-serif;
            background: var(--bg);
            color: var(--text);
            min-height: 100%;
            display: flex;
            align-items: center;
            justify-content: center;
            padding: 3rem 1.5rem 4rem 1.5rem;
        }

        .container {
            width: 100%;
            max-width: 380px;
        }

        /* Logo */
        .logo {
            display: flex;
            justify-content: center;
            margin-bottom: 1.25rem;
        }

        .logo svg {
            height: 24px;
            width: auto;
        }

        /* Badge - centered */
        .badge-wrapper {
            display: flex;
            justify-content: center;
            margin-bottom: 1.5rem;
        }

        .cli-badge {
            display: inline-flex;
            align-items: center;
            gap: 0.375rem;
            background: var(--primary-light);
            color: var(--primary);
            font-size: 0.75rem;
            font-weight: 600;
            padding: 0.375rem 0.75rem;
            border-radius: 100px;
        }

        .cli-badge svg {
            width: 14px;
            height: 14px;
        }

        h1 {
            font-size: 1.5rem;
            font-weight: 700;
            letter-spacing: -0.02em;
            margin-bottom: 0.375rem;
            text-align: center;
        }

        .subtitle {
            color: var(--text-secondary);
            font-size: 0.9375rem;
            margin-bottom: 1.25rem;
            text-align: center;
        }

        /* Credentials hint card */
        .credentials-hint {
            background: var(--bg-hint);
            border-radius: 12px;
            padding: 1rem;
            margin-bottom: 1rem;
        }

        .hint-header {
            display: flex;
            align-items: center;
            gap: 0.5rem;
            font-size: 0.8125rem;
            font-weight: 500;
            color: var(--text-secondary);
            margin-bottom: 0.75rem;
        }

        .hint-header svg {
            width: 16px;
            height: 16px;
            color: var(--text-muted);
        }

        .hint-links {
            display: flex;
            flex-direction: column;
            gap: 0.5rem;
        }

        .hint-link {
            display: flex;
            align-items: center;
            gap: 0.625rem;
            padding: 0.625rem 0.875rem;
            background: var(--bg-card);
            border: 1px solid var(--border);
            border-radius: 10px;
            text-decoration: none;
            color: var(--text);
            transition: all 0.15s ease;
        }

        .hint-link:hover {
            border-color: var(--primary);
            box-shadow: 0 0 0 2px rgba(124, 58, 237, 0.08);
        }

        .hint-link-icon {
            width: 32px;
            height: 32px;
            border-radius: 8px;
            display: flex;
            align-items: center;
            justify-content: center;
            flex-shrink: 0;
        }

        .hint-link-icon.production {
            background: var(--primary-light);
            color: var(--primary);
        }

        .hint-link-icon.sandbox {
            background: var(--coral-light);
            color: var(--coral);
        }

        .hint-link-icon svg {
            width: 16px;
            height: 16px;
        }

        .hint-link-text {
            flex: 1;
            min-width: 0;
        }

        .hint-link-title {
            font-weight: 600;
            font-size: 0.875rem;
        }

        .hint-link-path {
            font-size: 0.75rem;
            color: var(--text-muted);
        }

        .hint-link-arrow {
            color: var(--text-muted);
            flex-shrink: 0;
        }

        /* Form card */
        .form-card {
            background: var(--bg-card);
            border: 1px solid var(--border);
            border-radius: 12px;
            padding: 1.25rem;
            box-shadow: 0 1px 3px rgba(0, 0, 0, 0.04);
        }

        .form-group {
            margin-bottom: 1rem;
        }

        .form-group:last-of-type {
            margin-bottom: 0;
        }

        .label-row {
            display: flex;
            align-items: center;
            justify-content: space-between;
            margin-bottom: 0.375rem;
        }

        label {
            font-size: 0.875rem;
            font-weight: 600;
            color: var(--text);
        }

        .badge {
            font-size: 0.625rem;
            font-weight: 600;
            text-transform: uppercase;
            letter-spacing: 0.04em;
            padding: 0.1875rem 0.5rem;
            border-radius: 4px;
        }

        .badge-required {
            background: var(--primary-light);
            color: var(--primary);
        }

        .badge-optional {
            background: var(--coral-light);
            color: var(--coral);
        }

        .input-wrapper {
            position: relative;
        }

        input {
            width: 100%;
            padding: 0.625rem 0.875rem;
            font-family: inherit;
            font-size: 0.875rem;
            background: var(--bg-input);
            border: 1.5px solid transparent;
            border-radius: 8px;
            color: var(--text);
            transition: all 0.15s ease;
        }

        input::placeholder {
            color: var(--text-muted);
        }

        input:focus {
            outline: none;
            background: var(--bg-card);
            border-color: var(--primary);
            box-shadow: 0 0 0 3px rgba(124, 58, 237, 0.1);
        }

        input.mono {
            font-family: 'JetBrains Mono', monospace;
            font-size: 0.8125rem;
        }

        input.error {
            border-color: var(--error);
            background: var(--error-light);
        }

        input.error:focus {
            border-color: var(--error);
            box-shadow: 0 0 0 3px rgba(239, 68, 68, 0.1);
        }

        .input-hint {
            font-size: 0.75rem;
            color: var(--text-muted);
            margin-top: 0.25rem;
        }

        /* Password toggle */
        .password-toggle {
            position: absolute;
            right: 0.5rem;
            top: 50%;
            transform: translateY(-50%);
            background: none;
            border: none;
            color: var(--text-muted);
            cursor: pointer;
            padding: 0.25rem;
            border-radius: 4px;
            display: flex;
            align-items: center;
            justify-content: center;
        }

        .password-toggle:hover {
            color: var(--text-secondary);
        }

        .password-toggle svg {
            width: 18px;
            height: 18px;
        }

        /* Divider */
        .divider {
            display: flex;
            align-items: center;
            gap: 0.75rem;
            margin: 1rem 0;
            color: var(--text-muted);
            font-size: 0.6875rem;
            font-weight: 500;
            text-transform: uppercase;
            letter-spacing: 0.06em;
        }

        .divider::before,
        .divider::after {
            content: '';
            flex: 1;
            height: 1px;
            background: var(--border);
        }

        /* Buttons */
        .btn-group {
            display: flex;
            gap: 0.625rem;
            margin-top: 1.25rem;
        }

        button {
            flex: 1;
            padding: 0.6875rem 1rem;
            font-family: inherit;
            font-size: 0.875rem;
            font-weight: 600;
            border-radius: 8px;
            cursor: pointer;
            transition: all 0.15s ease;
            border: none;
        }

        .btn-secondary {
            background: var(--bg-input);
            color: var(--text-secondary);
            border: 1px solid var(--border);
        }

        .btn-secondary:hover {
            background: var(--border);
            color: var(--text);
        }

        .btn-primary {
            background: var(--primary);
            color: white;
            box-shadow: 0 2px 8px rgba(124, 58, 237, 0.25);
        }

        .btn-primary:hover {
            background: #6D28D9;
            box-shadow: 0 4px 12px rgba(124, 58, 237, 0.3);
        }

        button:disabled {
            opacity: 0.5;
            cursor: not-allowed;
        }

        /* Status - fixed toast above github link */
        .status {
            position: fixed;
            bottom: 3.5rem;
            left: 50%;
            transform: translateX(-50%) translateY(10px);
            padding: 0.625rem 1rem;
            border-radius: 8px;
            font-size: 0.8125rem;
            font-weight: 500;
            align-items: center;
            gap: 0.5rem;
            opacity: 0;
            visibility: hidden;
            transition: all 0.2s ease;
            display: flex;
            box-shadow: 0 4px 12px rgba(0, 0, 0, 0.1);
            z-index: 100;
            white-space: nowrap;
        }

        .status.show {
            opacity: 1;
            visibility: visible;
            transform: translateX(-50%) translateY(0);
        }
        .status.loading { background: var(--primary-light); color: var(--primary); }
        .status.success { background: var(--success-light); color: var(--success); }
        .status.error { background: var(--error-light); color: var(--error); }

        .spinner {
            width: 14px;
            height: 14px;
            border: 2px solid currentColor;
            border-top-color: transparent;
            border-radius: 50%;
            animation: spin 0.6s linear infinite;
        }

        @keyframes spin { to { transform: rotate(360deg); } }

        .status-icon { width: 14px; height: 14px; flex-shrink: 0; }

        .github-link {
            position: fixed;
            bottom: 1.5rem;
            left: 50%;
            transform: translateX(-50%);
            display: inline-flex;
            align-items: center;
            gap: 0.5rem;
            text-decoration: none;
            color: #9CA3AF;
            font-size: 0.75rem;
            font-weight: 500;
            letter-spacing: 0.01em;
            transition: color 0.2s ease;
        }

        .github-link:hover {
            color: #6B7280;
        }

        .github-link:hover .github-icon {
            color: #6B7280;
        }

        .github-icon {
            width: 16px;
            height: 16px;
            flex-shrink: 0;
            color: #9CA3AF;
            transition: color 0.2s ease;
        }
    </style>
</head>
<body>
    <div class="container">
        <div class="logo">
            <svg viewBox="0 0 584.1 79.9" xmlns="http://www.w3.org/2000/svg">
                <defs>
                    <linearGradient id="SVGID_1_" gradientUnits="userSpaceOnUse" x1="0" y1="42.9557" x2="120.102" y2="42.9557" gradientTransform="matrix(1 0 0 -1 0 82.8742)">
                        <stop offset="0" style="stop-color:#EF4749"/>
                        <stop offset="1" style="stop-color:#F68D41"/>
                    </linearGradient>
                </defs>
                <g fill="#080808">
                    <path d="M211.4,77.2L183.9,3.7c-0.1-0.3-0.4-0.6-0.8-0.6h-18.3c-0.4,0-0.7,0.2-0.8,0.5l-27.8,73.5c-0.2,0.6,0.2,1.1,0.8,1.1h15.9c0.4,0,0.7-0.2,0.8-0.6l5-14.1c0.1-0.3,0.4-0.6,0.8-0.6h28c0.4,0,0.7,0.2,0.8,0.6l5,14.1c0.1,0.3,0.4,0.6,0.8,0.6h16.5C211.2,78.4,211.6,77.8,211.4,77.2z M182.1,47.8H165c-0.3,0-0.5-0.3-0.4-0.6l8.7-24.3c0.1-0.4,0.7-0.4,0.8,0l8.5,24.3C182.6,47.5,182.4,47.8,182.1,47.8z"/>
                    <path d="M224.6,0.1c5.3,0,9.6,4.2,9.6,9.4s-4.3,9.4-9.6,9.4c-5.1,0-9.4-4.2-9.4-9.4S219.5,0.1,224.6,0.1z"/>
                    <path d="M216.6,77.5v-51c0-0.5,0.4-0.8,0.9-0.8H232c0.5,0,0.9,0.4,0.9,0.8v51c0,0.5-0.4,0.8-0.9,0.8h-14.5C217,78.4,216.6,78,216.6,77.5z"/>
                    <path d="M274.5,40.5c0,0.5-0.5,0.9-1,0.8c-1.4-0.3-2.8-0.3-4.1-0.3c-6.5,0-12.4,3.8-12.4,14.3v22.2c0,0.5-0.4,0.8-0.9,0.8h-14.5c-0.5,0-0.9-0.4-0.9-0.8v-51c0-0.5,0.4-0.8,0.9-0.8h14c0.5,0,0.9,0.4,0.9,0.8v6.3c2.8-5.9,9.5-7.6,13.8-7.6c1.3,0,2.6,0.1,3.6,0.4c0.4,0.1,0.7,0.4,0.7,0.8v14.1H274.5z"/>
                    <path d="M332.1,26.2l9.7,29.4c0.1,0.4,0.7,0.4,0.8,0l8.3-29.3c0.1-0.4,0.4-0.6,0.8-0.6h14.1c0.6,0,1,0.6,0.8,1.1l-15.9,51c-0.1,0.4-0.4,0.6-0.8,0.6H335c-0.4,0-0.7-0.2-0.8-0.6l-10.9-32.3c-0.1-0.4-0.7-0.4-0.8,0l-10.7,32.3c-0.1,0.3-0.4,0.6-0.8,0.6h-15.2c-0.4,0-0.7-0.2-0.8-0.6l-16.1-51c-0.2-0.5,0.2-1.1,0.8-1.1h15c0.4,0,0.7,0.3,0.8,0.6l8.3,29.2c0.1,0.4,0.7,0.4,0.8,0l9.8-29.3c0.1-0.3,0.4-0.6,0.8-0.6h16.1C331.6,25.6,331.9,25.9,332.1,26.2z"/>
                    <path d="M422,25.6h-14c-0.5,0-0.9,0.4-0.9,0.8v5.2c-1.1-2.4-5.2-7.6-14.7-7.6c-15.7,0-25.8,12.1-25.8,27.8c0,16.2,10.8,28,26.3,28c6.6,0,11.9-3.2,14.3-8.3c0,0.3-0.1,4.2-0.1,5.9c0,0.5,0.4,0.9,0.9,0.9h14c0.5,0,0.9-0.4,0.9-0.8v-51C422.9,26,422.5,25.6,422,25.6z M395.1,65.9c-6.7,0-12.2-5.3-12.2-14c0-9.1,5.2-13.9,12.2-13.9c6.6,0,12.1,4.8,12.1,13.9C407.2,60.8,401.6,65.9,395.1,65.9z"/>
                    <path d="M431,77.5V2.4c0-0.5,0.4-0.8,0.9-0.8h14.5c0.5,0,0.9,0.4,0.9,0.8v75.1c0,0.5-0.4,0.8-0.9,0.8h-14.5C431.4,78.4,431,78,431,77.5z"/>
                    <path d="M454.9,77.5V2.4c0-0.5,0.4-0.8,0.9-0.8h14.5c0.5,0,0.9,0.4,0.9,0.8v75.1c0,0.5-0.4,0.8-0.9,0.8h-14.5C455.3,78.4,454.9,78,454.9,77.5z"/>
                    <path d="M546.4,51.6L529.1,27c-0.4-0.6,0-1.3,0.7-1.3h17c0.3,0,0.6,0.1,0.7,0.4l8.9,13.5c0.2,0.3,0.5,0.3,0.7,0l8.7-13.5c0.2-0.2,0.4-0.4,0.7-0.4h16c0.7,0,1.1,0.8,0.7,1.3l-17,24c-0.2,0.3-0.2,0.7,0,1c5.6,7.9,12,17.1,17.7,25.1c0.4,0.6,0,1.3-0.7,1.3h-16.9c-0.3,0-0.6-0.1-0.7-0.4l-9.2-13.9c-0.2-0.3-0.5-0.3-0.7,0c-2.8,4.1-6.4,9.8-9.1,13.9c-0.2,0.2-0.4,0.4-0.7,0.4h-15.7c-0.7,0-1.1-0.8-0.7-1.3l17-24.5C546.6,52.2,546.6,51.9,546.4,51.6z"/>
                    <path d="M492.7,56.1h18.5h18c0.1-0.3,0.3-2.8,0.3-5c0-17-10.1-27.1-26.6-27.1c-13.8,0-26.4,10.8-26.4,27.8c0,17.7,13,28.1,27.6,28.1c13.2,0,21.6-7.4,24.3-16.3c0-0.1,0.1-0.4,0.2-0.9c0.1-0.5-0.3-1-0.8-1h-13.5c-0.3,0-0.5,0.1-0.7,0.4c-1.8,2.6-4.8,4.2-9.3,4.2c-6.1,0-11.4-4-12-9.7C492.2,56.3,492.4,56.1,492.7,56.1z M503.1,36.8c7.4,0,10.2,4.5,10.6,8.4c0,0.3-0.2,0.5-0.4,0.5H493c-0.3,0-0.5-0.2-0.4-0.5C493.1,41.4,496.5,36.8,503.1,36.8z"/>
                </g>
                <path fill="url(#SVGID_1_)" d="M115.5,31.4c-4.3-4.2-10.3-5.4-15.8-3.1L74.9,38.4l-27-32.3C44,1.4,38-0.8,32,0.2c-6,1.1-10.9,5.1-13,10.9L0.9,59.9c-2.4,6.4-0.1,13.7,5.7,17.5c4.1,2.7,9.3,3,13.8,1.2l24.5-10c4.1-1.7,6.4-6.3,5-10.5c-1.5-4.6-6.6-7-11-5.2l-18.6,7.7c-0.8,0.3-1.7-0.5-1.4-1.3L34.2,18c0.3-0.7,1.3-0.9,1.8-0.3l47,56.2c3.2,3.8,7.8,5.9,12.6,5.9c1.1,0,2.1-0.1,3.2-0.3c5.8-1.1,10.5-5.4,12.5-11l7.8-21.3C121.2,41.6,119.8,35.5,115.5,31.4z M101.2,47.6l-5.1,13.8c-0.3,0.8-1.3,0.9-1.8,0.3l-8.2-9.8l13.7-5.6C100.7,45.9,101.6,46.7,101.2,47.6z"/>
            </svg>
        </div>

        <div class="badge-wrapper">
            <div class="cli-badge">
            <svg viewBox="0 0 16 16" fill="none" xmlns="http://www.w3.org/2000/svg">
                <rect x="2" y="2" width="12" height="12" rx="2" stroke="currentColor" stroke-width="1.5"/>
                <path d="M5 6L7 8L5 10" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round"/>
                <path d="M9 10H11" stroke="currentColor" stroke-width="1.5" stroke-linecap="round"/>
            </svg>
            CLI Authentication
            </div>
        </div>

        <div class="credentials-hint">
            <div class="hint-header">
                <svg viewBox="0 0 16 16" fill="none" xmlns="http://www.w3.org/2000/svg">
                    <circle cx="8" cy="8" r="6.5" stroke="currentColor" stroke-width="1.5"/>
                    <path d="M8 7V11" stroke="currentColor" stroke-width="1.5" stroke-linecap="round"/>
                    <circle cx="8" cy="4.75" r="0.75" fill="currentColor"/>
                </svg>
                Where to find your API credentials
            </div>
            <div class="hint-links">
                <a href="https://www.airwallex.com/app/settings/developer" target="_blank" class="hint-link">
                    <div class="hint-link-icon production">
                        <svg viewBox="0 0 16 16" fill="none" xmlns="http://www.w3.org/2000/svg">
                            <path d="M8 2L2 5.5V10.5L8 14L14 10.5V5.5L8 2Z" stroke="currentColor" stroke-width="1.5" stroke-linejoin="round"/>
                            <path d="M2 5.5L8 9M8 9L14 5.5M8 9V14" stroke="currentColor" stroke-width="1.5" stroke-linejoin="round"/>
                        </svg>
                    </div>
                    <div class="hint-link-text">
                        <div class="hint-link-title">Production API Keys</div>
                        <div class="hint-link-path">airwallex.com/app &rarr; Settings &rarr; Developer</div>
                    </div>
                    <svg class="hint-link-arrow" width="16" height="16" viewBox="0 0 16 16" fill="none">
                        <path d="M6 4L10 8L6 12" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round"/>
                    </svg>
                </a>
                <a href="https://demo.airwallex.com/app/settings/developer" target="_blank" class="hint-link">
                    <div class="hint-link-icon sandbox">
                        <svg viewBox="0 0 16 16" fill="none" xmlns="http://www.w3.org/2000/svg">
                            <path d="M5.5 6.5L3.5 8L5.5 9.5M10.5 6.5L12.5 8L10.5 9.5" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round"/>
                            <path d="M9 4L7 12" stroke="currentColor" stroke-width="1.5" stroke-linecap="round"/>
                        </svg>
                    </div>
                    <div class="hint-link-text">
                        <div class="hint-link-title">Sandbox API Keys</div>
                        <div class="hint-link-path">demo.airwallex.com/app &rarr; Settings &rarr; Developer</div>
                    </div>
                    <svg class="hint-link-arrow" width="16" height="16" viewBox="0 0 16 16" fill="none">
                        <path d="M6 4L10 8L6 12" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round"/>
                    </svg>
                </a>
            </div>
        </div>

        <div class="form-card">
            <form id="setupForm" autocomplete="off">
                <div class="form-group">
                    <div class="label-row">
                        <label for="accountName">Account Name</label>
                        <span class="badge badge-required">Required</span>
                    </div>
                    <input type="text" id="accountName" name="accountName" class="mono" placeholder="e.g., production, sandbox" required autofocus>
                    <div class="input-hint">A local identifier for this account configuration</div>
                </div>

                <div class="form-group">
                    <div class="label-row">
                        <label for="clientId">Client ID</label>
                        <span class="badge badge-required">Required</span>
                    </div>
                    <input type="text" id="clientId" name="clientId" class="mono" placeholder="Your Airwallex Client ID" required>
                </div>

                <div class="form-group">
                    <div class="label-row">
                        <label for="apiKey">API Key</label>
                        <span class="badge badge-required">Required</span>
                    </div>
                    <div class="input-wrapper">
                        <input type="password" id="apiKey" name="apiKey" class="mono" placeholder="Your Airwallex API Key" required style="padding-right: 2.25rem;">
                        <button type="button" class="password-toggle" id="togglePassword" aria-label="Toggle password visibility">
                            <svg id="eyeIcon" viewBox="0 0 18 18" fill="none">
                                <path d="M2 9C2 9 4.5 4 9 4C13.5 4 16 9 16 9C16 9 13.5 14 9 14C4.5 14 2 9 2 9Z" stroke="currentColor" stroke-width="1.5" stroke-linecap="round"/>
                                <circle cx="9" cy="9" r="2" stroke="currentColor" stroke-width="1.5"/>
                            </svg>
                            <svg id="eyeOffIcon" style="display:none" viewBox="0 0 18 18" fill="none">
                                <path d="M7.6 7.6a2 2 0 1 0 2.8 2.8M12.5 12.5A6.5 6.5 0 0 1 9 14c-4.5 0-7-5-7-5a11.5 11.5 0 0 1 3-3.5m2.2-1.2A5.5 5.5 0 0 1 9 4c4.5 0 7 5 7 5a11.5 11.5 0 0 1-1.2 1.8" stroke="currentColor" stroke-width="1.5" stroke-linecap="round"/>
                                <path d="M2 2l14 14" stroke="currentColor" stroke-width="1.5" stroke-linecap="round"/>
                            </svg>
                        </button>
                    </div>
                </div>

                <div class="divider">Multi-Account</div>

                <div class="form-group">
                    <div class="label-row">
                        <label for="accountId">Account ID</label>
                        <span class="badge badge-optional">Optional</span>
                    </div>
                    <input type="text" id="accountId" name="accountId" class="mono" placeholder="acct_xxxxxxxxxxxxxxxxxx">
                    <div class="input-hint">Required if your API key has access to multiple accounts</div>
                </div>

                <div class="btn-group">
                    <button type="button" id="testBtn" class="btn-secondary">Test Connection</button>
                    <button type="submit" id="submitBtn" class="btn-primary">Save & Connect</button>
                </div>

                <div id="status" class="status"></div>
            </form>
        </div>
    </div>

    <script>
        const form = document.getElementById('setupForm');
        const testBtn = document.getElementById('testBtn');
        const submitBtn = document.getElementById('submitBtn');
        const status = document.getElementById('status');
        const togglePassword = document.getElementById('togglePassword');
        const apiKeyInput = document.getElementById('apiKey');
        const eyeIcon = document.getElementById('eyeIcon');
        const eyeOffIcon = document.getElementById('eyeOffIcon');
        const csrfToken = '{{.CSRFToken}}';

        const requiredFields = ['accountName', 'clientId', 'apiKey'];
        let isBusy = false;

        // Clear error state when user types
        requiredFields.forEach(id => {
            document.getElementById(id).addEventListener('input', function() {
                this.classList.remove('error');
            });
        });

        togglePassword.addEventListener('click', () => {
            const isPassword = apiKeyInput.type === 'password';
            apiKeyInput.type = isPassword ? 'text' : 'password';
            eyeIcon.style.display = isPassword ? 'none' : 'block';
            eyeOffIcon.style.display = isPassword ? 'block' : 'none';
        });

        function showStatus(type, message) {
            status.className = 'status show ' + type;
            if (type === 'loading') {
                status.innerHTML = '<div class="spinner"></div><span>' + message + '</span>';
            } else {
                const icon = type === 'success'
                    ? '<svg class="status-icon" viewBox="0 0 16 16" fill="none"><path d="M13 5L6.5 11.5L3 8" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"/></svg>'
                    : '<svg class="status-icon" viewBox="0 0 16 16" fill="none"><path d="M12 4L4 12M4 4L12 12" stroke="currentColor" stroke-width="2" stroke-linecap="round"/></svg>';
                status.innerHTML = icon + '<span>' + message + '</span>';
            }
        }

        function hideStatus() {
            status.className = 'status';
        }

        function validateRequired() {
            let valid = true;
            requiredFields.forEach(id => {
                const input = document.getElementById(id);
                if (!input.value.trim()) {
                    input.classList.add('error');
                    valid = false;
                }
            });
            return valid;
        }

        function getFormData() {
            return {
                account_name: document.getElementById('accountName').value.trim(),
                client_id: document.getElementById('clientId').value.trim(),
                api_key: document.getElementById('apiKey').value.trim(),
                account_id: document.getElementById('accountId').value.trim()
            };
        }

        testBtn.addEventListener('click', async () => {
            if (isBusy) return;
            isBusy = true;
            hideStatus();
            if (!validateRequired()) {
                isBusy = false;
                return;
            }

            const data = getFormData();
            testBtn.disabled = true;
            submitBtn.disabled = true;
            showStatus('loading', 'Testing connection...');
            try {
                const response = await fetch('/validate', {
                    method: 'POST',
                    headers: { 'Content-Type': 'application/json', 'X-CSRF-Token': csrfToken },
                    body: JSON.stringify(data)
                });
                const result = await response.json();
                showStatus(result.success ? 'success' : 'error', result.success ? 'Connection successful!' : result.error);
            } catch (err) {
                showStatus('error', 'Request failed: ' + err.message);
            } finally {
                testBtn.disabled = false;
                submitBtn.disabled = false;
                isBusy = false;
            }
        });

        form.addEventListener('submit', async (e) => {
            e.preventDefault();
            if (isBusy) return;
            isBusy = true;
            hideStatus();
            if (!validateRequired()) {
                isBusy = false;
                return;
            }

            const data = getFormData();
            testBtn.disabled = true;
            submitBtn.disabled = true;
            showStatus('loading', 'Saving credentials...');
            try {
                const response = await fetch('/submit', {
                    method: 'POST',
                    headers: { 'Content-Type': 'application/json', 'X-CSRF-Token': csrfToken },
                    body: JSON.stringify(data)
                });
                const result = await response.json();
                if (result.success) {
                    showStatus('success', 'Credentials saved! Redirecting...');
                    setTimeout(() => { window.location.href = '/success'; }, 600);
                } else {
                    showStatus('error', result.error);
                    testBtn.disabled = false;
                    submitBtn.disabled = false;
                    isBusy = false;
                }
            } catch (err) {
                showStatus('error', 'Request failed: ' + err.message);
                testBtn.disabled = false;
                submitBtn.disabled = false;
                isBusy = false;
            }
        });
    </script>

    <a href="https://github.com/salmonumbrella/airwallex-cli" target="_blank" class="github-link">
        <svg class="github-icon" viewBox="0 0 16 16" fill="currentColor">
            <path d="M8 0C3.58 0 0 3.58 0 8c0 3.54 2.29 6.53 5.47 7.59.4.07.55-.17.55-.38 0-.19-.01-.82-.01-1.49-2.01.37-2.53-.49-2.69-.94-.09-.23-.48-.94-.82-1.13-.28-.15-.68-.52-.01-.53.63-.01 1.08.58 1.23.82.72 1.21 1.87.87 2.33.66.07-.52.28-.87.51-1.07-1.78-.2-3.64-.89-3.64-3.95 0-.87.31-1.59.82-2.15-.08-.2-.36-1.02.08-2.12 0 0 .67-.21 2.2.82.64-.18 1.32-.27 2-.27.68 0 1.36.09 2 .27 1.53-1.04 2.2-.82 2.2-.82.44 1.1.16 1.92.08 2.12.51.56.82 1.27.82 2.15 0 3.07-1.87 3.75-3.65 3.95.29.25.54.73.54 1.48 0 1.07-.01 1.93-.01 2.2 0 .21.15.46.55.38A8.013 8.013 0 0016 8c0-4.42-3.58-8-8-8z"/>
        </svg>
        Airwallex CLI
    </a>
</body>
</html>`

const successTemplate = `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0, maximum-scale=1.0, user-scalable=no">
    <title>Connected - Airwallex CLI</title>
    <link rel="preconnect" href="https://fonts.googleapis.com">
    <link rel="preconnect" href="https://fonts.gstatic.com" crossorigin>
    <link href="https://fonts.googleapis.com/css2?family=Inter:wght@400;500;600;700&family=JetBrains+Mono:wght@400;500&display=swap" rel="stylesheet">
    <style>
        :root {
            --bg: #F8F9FB;
            --bg-card: #FFFFFF;
            --bg-terminal: #1F2937;
            --border: #E5E7EB;
            --text: #1F2937;
            --text-secondary: #6B7280;
            --text-muted: #9CA3AF;
            --text-dim: #D1D5DB;
            --primary: #7C3AED;
            --primary-light: #EDE9FE;
            --success: #10B981;
            --success-light: #D1FAE5;
        }

        * { margin: 0; padding: 0; box-sizing: border-box; }
        html {
            height: 100%;
        }

        body {
            font-family: 'Inter', -apple-system, sans-serif;
            background: var(--bg);
            color: var(--text);
            min-height: 100%;
            display: flex;
            flex-direction: column;
            align-items: center;
            justify-content: center;
            padding: 3rem 1.5rem 4rem 1.5rem;
        }

        .container {
            width: 100%;
            max-width: 380px;
            text-align: center;
        }

        .success-icon {
            width: 56px;
            height: 56px;
            background: var(--success-light);
            border-radius: 50%;
            margin: 0 auto 1.25rem;
            display: flex;
            align-items: center;
            justify-content: center;
            animation: scaleIn 0.5s cubic-bezier(0.34, 1.56, 0.64, 1) forwards;
        }

        @keyframes scaleIn {
            from { transform: scale(0); }
            to { transform: scale(1); }
        }

        .success-icon svg {
            width: 28px;
            height: 28px;
            color: var(--success);
        }

        h1 {
            font-size: 1.375rem;
            font-weight: 700;
            margin-bottom: 0.25rem;
            animation: fadeUp 0.5s ease 0.2s both;
        }

        .subtitle {
            color: var(--text-secondary);
            font-size: 0.875rem;
            margin-bottom: 1rem;
            animation: fadeUp 0.5s ease 0.3s both;
        }

        @keyframes fadeUp {
            from { opacity: 0; transform: translateY(10px); }
            to { opacity: 1; transform: translateY(0); }
        }

        .account-badge {
            display: inline-flex;
            align-items: center;
            gap: 0.375rem;
            background: var(--primary-light);
            color: var(--primary);
            font-size: 0.8125rem;
            font-weight: 600;
            padding: 0.375rem 0.875rem;
            border-radius: 100px;
            margin-bottom: 1.25rem;
            animation: fadeUp 0.5s ease 0.35s both;
        }

        .account-badge .dot {
            width: 6px;
            height: 6px;
            background: var(--success);
            border-radius: 50%;
            animation: dotPulse 2s ease-in-out infinite;
        }

        @keyframes dotPulse {
            0%, 100% { opacity: 1; }
            50% { opacity: 0.5; }
        }

        .terminal {
            background: var(--bg-terminal);
            border-radius: 10px;
            overflow: hidden;
            text-align: left;
            animation: fadeUp 0.5s ease 0.4s both;
            box-shadow: 0 4px 20px rgba(0, 0, 0, 0.12);
        }

        .terminal-bar {
            background: #111827;
            padding: 0.5rem 0.75rem;
            display: flex;
            align-items: center;
            gap: 0.375rem;
        }

        .terminal-dot {
            width: 9px;
            height: 9px;
            border-radius: 50%;
        }

        .terminal-dot.red { background: #FF5F57; }
        .terminal-dot.yellow { background: #FEBC2E; }
        .terminal-dot.green { background: #28C840; }

        .terminal-body {
            padding: 0.875rem;
        }

        .terminal-line {
            display: flex;
            align-items: center;
            gap: 0.5rem;
            font-family: 'JetBrains Mono', monospace;
            font-size: 0.75rem;
            margin-bottom: 0.5rem;
            color: #E5E7EB;
        }

        .terminal-line:last-child { margin-bottom: 0; }
        .terminal-prompt { color: var(--primary); user-select: none; }
        .terminal-cmd { color: #10B981; }
        .terminal-output { color: #9CA3AF; padding-left: 1.125rem; margin-top: -0.25rem; margin-bottom: 0.5rem; font-size: 0.6875rem; }

        .terminal-cursor {
            display: inline-block;
            width: 9px;
            height: 16px;
            background: var(--primary);
            animation: cursorBlink 1.2s step-end infinite;
            margin-left: 2px;
            vertical-align: middle;
        }

        @keyframes cursorBlink {
            0%, 50% { opacity: 1; }
            50.01%, 100% { opacity: 0; }
        }

        .message {
            margin-top: 1rem;
            padding: 0.875rem;
            background: var(--primary-light);
            border: 1px solid rgba(124, 58, 237, 0.12);
            border-radius: 8px;
            text-align: center;
            animation: fadeUp 0.5s ease 0.5s both;
        }

        .message-icon {
            font-size: 1.125rem;
            margin-bottom: 0.125rem;
            color: var(--primary);
        }

        .message-title {
            font-weight: 600;
            font-size: 0.875rem;
            margin-bottom: 0.125rem;
            color: var(--text);
        }

        .message-text {
            font-size: 0.75rem;
            color: var(--text-secondary);
            line-height: 1.4;
        }

        .message-text code {
            font-family: 'JetBrains Mono', monospace;
            background: var(--bg-card);
            color: var(--primary);
            padding: 0.0625rem 0.3125rem;
            border-radius: 3px;
            font-size: 0.6875rem;
        }

        .footer {
            margin-top: 1rem;
            font-size: 0.75rem;
            color: var(--text-muted);
            animation: fadeUp 0.5s ease 0.6s both;
        }

        .github-link {
            position: fixed;
            bottom: 1.5rem;
            left: 50%;
            transform: translateX(-50%);
            display: inline-flex;
            align-items: center;
            gap: 0.5rem;
            text-decoration: none;
            color: #9CA3AF;
            font-size: 0.75rem;
            font-weight: 500;
            letter-spacing: 0.01em;
            transition: color 0.2s ease;
            animation: fadeUp 0.5s ease 0.7s both;
        }

        .github-link:hover {
            color: #6B7280;
        }

        .github-link:hover .github-icon {
            color: #6B7280;
        }

        .github-icon {
            width: 16px;
            height: 16px;
            flex-shrink: 0;
            color: #9CA3AF;
            transition: color 0.2s ease;
        }
    </style>
</head>
<body>
    <div class="container">
        <div class="success-icon">
            <svg viewBox="0 0 32 32" fill="none">
                <path d="M8 16L14 22L24 10" stroke="currentColor" stroke-width="3" stroke-linecap="round" stroke-linejoin="round"/>
            </svg>
        </div>

        <h1>You're all set!</h1>
        <p class="subtitle">Airwallex CLI is now connected and ready to use</p>

        {{if .AccountName}}
        <div class="account-badge">
            <span class="dot"></span>
            <span>{{.AccountName}}</span>
        </div>
        {{end}}

        <div class="terminal">
            <div class="terminal-bar">
                <span class="terminal-dot red"></span>
                <span class="terminal-dot yellow"></span>
                <span class="terminal-dot green"></span>
            </div>
            <div class="terminal-body">
                <div class="terminal-line">
                    <span class="terminal-prompt">$</span>
                    <span class="terminal-cmd">airwallex</span>
                    <span>balances</span>
                </div>
                <div class="terminal-output">Fetching account balances...</div>
                <div class="terminal-line">
                    <span class="terminal-prompt">$</span>
                    <span class="terminal-cmd">airwallex</span>
                    <span>transfers list</span>
                </div>
                <div class="terminal-output">Listing recent transfers...</div>
                <div class="terminal-line">
                    <span class="terminal-prompt">$</span>
                    <span class="terminal-cursor"></span>
                </div>
            </div>
        </div>

        <div class="message">
            <div class="message-icon">&larr;</div>
            <div class="message-title">Return to your terminal</div>
            <div class="message-text">You can close this window and start using the CLI. Try running <code>airwallex --help</code> to see all available commands.</div>
        </div>

        <p class="footer">This window will close automatically.</p>
    </div>

    <a href="https://github.com/salmonumbrella/airwallex-cli" target="_blank" class="github-link">
        <svg class="github-icon" viewBox="0 0 16 16" fill="currentColor">
            <path d="M8 0C3.58 0 0 3.58 0 8c0 3.54 2.29 6.53 5.47 7.59.4.07.55-.17.55-.38 0-.19-.01-.82-.01-1.49-2.01.37-2.53-.49-2.69-.94-.09-.23-.48-.94-.82-1.13-.28-.15-.68-.52-.01-.53.63-.01 1.08.58 1.23.82.72 1.21 1.87.87 2.33.66.07-.52.28-.87.51-1.07-1.78-.2-3.64-.89-3.64-3.95 0-.87.31-1.59.82-2.15-.08-.2-.36-1.02.08-2.12 0 0 .67-.21 2.2.82.64-.18 1.32-.27 2-.27.68 0 1.36.09 2 .27 1.53-1.04 2.2-.82 2.2-.82.44 1.1.16 1.92.08 2.12.51.56.82 1.27.82 2.15 0 3.07-1.87 3.75-3.65 3.95.29.25.54.73.54 1.48 0 1.07-.01 1.93-.01 2.2 0 .21.15.46.55.38A8.013 8.013 0 0016 8c0-4.42-3.58-8-8-8z"/>
        </svg>
        Airwallex CLI
    </a>

    <script>fetch('/complete', { method: 'POST', headers: { 'X-CSRF-Token': '{{.CSRFToken}}' } }).catch(() => {});</script>
</body>
</html>`
