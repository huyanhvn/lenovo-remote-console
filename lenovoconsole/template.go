package lenovoconsole

// htmlTemplate contains the HTML template for the console viewer
const htmlTemplate = `<!DOCTYPE html>
<html>
<head>
    <title>Lenovo XCC Remote Console</title>
    <style>
        body {
            margin: 0;
            padding: 0;
            background-color: #000;
            overflow: hidden;
            font-family: Arial, sans-serif;
        }
        #status {
            position: absolute;
            top: 10px;
            left: 10px;
            color: #fff;
            background: rgba(0,0,0,0.7);
            padding: 10px;
            border-radius: 5px;
            z-index: 1000;
            max-width: 500px;
        }
        #kvmCanvas {
            display: block;
            margin: 0 auto;
        }
        .error {
            color: #ff4444;
            font-weight: bold;
        }
        #certInstructions {
            position: absolute;
            top: 10px;
            right: 10px;
            color: #fff;
            background: rgba(0,0,50,0.9);
            padding: 15px;
            border-radius: 5px;
            z-index: 1001;
            max-width: 400px;
            border: 2px solid #4444ff;
            display: none;
        }
        #certInstructions h3 {
            margin-top: 0;
            color: #88aaff;
        }
        #certInstructions a {
            color: #88aaff;
            text-decoration: underline;
        }
        #certInstructions ol {
            padding-left: 20px;
        }
        #certInstructions li {
            margin-bottom: 10px;
        }
        #certInstructions button {
            background: #4444ff;
            color: white;
            border: none;
            padding: 8px 15px;
            border-radius: 3px;
            cursor: pointer;
            margin-top: 10px;
        }
        #certInstructions button:hover {
            background: #5555ff;
        }
    </style>
</head>
<body>
    <div id="status">Initializing console...</div>
    <canvas id="kvmCanvas"></canvas>
    
    <div id="certInstructions">
        <h3>⚠️ Certificate Issue Detected</h3>
        <p>The BMC server is using a self-signed certificate that needs to be accepted.</p>
        <p><strong>To fix this issue:</strong></p>
        <ol>
            <li>Click the button below to open the BMC certificate page</li>
            <li>You'll see a browser warning about the certificate</li>
            <li>Click "Advanced" or "Show Details"</li>
            <li>Click "Proceed to {{.BMCIP}}" or "Accept the Risk and Continue"</li>
            <li>Come back to this tab and click "Retry Connection"</li>
        </ol>
        <button onclick="acceptCertificate()">Open BMC Certificate Page</button>
        <button onclick="retryConnection()">Retry Connection</button>
        <button onclick="document.getElementById('certInstructions').style.display='none'">Close</button>
    </div>

    <script>
        // Console configuration
        const config = {
            bmcIP: '{{.BMCIP}}',
            rpPort: {{.RPPort}},
            bmcUsername: '{{.BMCUsername}}',
            bmcPassword: '{{.BMCPassword}}'
        };

        const statusDiv = document.getElementById('status');
        let scriptsLoaded = 0;
        const requiredScripts = [
            '/SDK_Pilot4/utility.js',
            '/SDK_Pilot4/rpimage.js', 
            '/SDK_Pilot4/rprecorder.js',
            '/SDK_Pilot4/rpviewer.js',
			'/SDK_Pilot4/rphandlers.js',
			'/SDK_Pilot4/websockethandler.js',
			'/SDK_Pilot4/virtualkeyboard.js',
			'/SDK_Pilot4/mediaTypes.js',
			'/SDK_Pilot4/mediaworkerhandler.js'
        ];

        function updateStatus(message, isError) {
            statusDiv.innerHTML = message;
            if (isError) {
                statusDiv.className = 'error';
            }
        }

        function loadScript(src, callback, errorCallback) {
            const script = document.createElement('script');
            script.src = 'https://' + config.bmcIP + src;
            script.onload = callback;
            script.onerror = errorCallback || function() {
                console.error('Failed to load:', src);
            };
            document.head.appendChild(script);
        }

        function loadNextScript(index) {
            if (index >= requiredScripts.length) {
                initializeViewer();
                return;
            }

            const scriptName = requiredScripts[index].split('/').pop();
            updateStatus('Loading ' + scriptName + '... (' + (index + 1) + '/' + requiredScripts.length + ')');

            loadScript(
                requiredScripts[index],
                function() {
                    console.log('Loaded:', requiredScripts[index]);
                    loadNextScript(index + 1);
                },
                function() {
                    // Try alternative path without /designs/imm
                    const altPath = requiredScripts[index].replace('/designs/imm', '');
                    console.log('Trying alternative path:', altPath);
                    
                    loadScript(
                        altPath,
                        function() {
                            console.log('Loaded from alt path:', altPath);
                            loadNextScript(index + 1);
                        },
                        function() {
                            updateStatus('❌ ERROR: Could not load ' + scriptName + '<br>' +
                                       'Tried paths:<br>' +
                                       '- https://' + config.bmcIP + requiredScripts[index] + '<br>' +
                                       '- https://' + config.bmcIP + altPath + '<br><br>' +
                                       'Please check browser console for details.', true);
                        }
                    );
                }
            );
        }

        updateStatus('⚠️ Loading Lenovo RPViewer libraries...');
        loadNextScript(0);

        // Certificate acceptance handler
        function handleCertificateAcceptance(viewer) {
            // Note: Direct iframe approach won't work due to CSP restrictions
            // The user will need to manually accept the certificate if prompted
            console.log('Certificate handling will be done through RPViewer dialog');
        }

        function initializeViewer() {
            updateStatus('✓ All libraries loaded. Initializing viewer...');
            
            try {
                // Check if RPViewer is available
                if (typeof RPViewer === 'undefined') {
                    throw new Error('RPViewer class not found. Check that all scripts loaded correctly.');
                }

                console.log('Initializing RPViewer...');
                
                // Initialize RPViewer (based on the RemoteConsoleWindow.js code)
                const viewer = new RPViewer('kvmCanvas', viewerAPIErrorCallback);
                
                console.log('RPViewer created, configuring...');
                
                // Configure viewer
                viewer.setRPWebSocketTimeout(30);
                
                // Certificate handling - try without setting cert file
                // Since the certificate is already accepted, we might not need this
                // viewer.setRPCertFileName('/cert.pem');
                
                // Set server configuration
                viewer.setRPServerConfiguration(config.bmcIP, config.rpPort);
                viewer.setRPEmbeddedViewerSize(window.innerWidth, window.innerHeight - 50);
                
                // Connection settings - Multi User Mode
                viewer.setRPExclusiveLogin(false); // Use multi-user mode (non-exclusive)
                viewer.setRPAllowSharingRequests(true); // Allow sharing requests for multi-user mode
                
                // Input support
                viewer.setRPMouseInputSupport(true);
                viewer.setRPTouchInputSupport(true);
                viewer.setRPKeyboardInputSupport(true);
                
                // Debug settings
                viewer.setRPDebugMode(true);
                viewer.setRPDebugLevel(1);
                
                // Display settings
                viewer.setRPMaintainAspectRatio(true);
                viewer.setRPInitialBackgroundColor('black');
                viewer.setRPInitialMessageColor('white');
                viewer.setRPKeyboardLanguage('en');
                
                // Reconnection settings
                viewer.setRPSupportReconnect(true);
                viewer.setRPLinkInterruptMessageColor('red');
                viewer.setRPLinkInterruptMessage('Connection interrupted. Attempting to reconnect...');
                viewer.setRPReconnectingMessage('Reconnecting to remote console...');
                viewer.setRPInitialMessage('Connecting to remote console...');
                
                // Set credentials (using BMC credentials directly)
                console.log('Setting credentials with BMC user:', config.bmcUsername);
                viewer.setRPCredential(config.bmcUsername, config.bmcPassword);
                
                // Register callbacks
                viewer.registerRPLoginResponseCallback(loginResponseCallback);
                viewer.registerRPUIInitCallback(uiInitCallback);
                viewer.registerRPExitViewerCallback(exitViewerCallback);
                viewer.registerRPResolutionCallback(resolutionCallback);
                viewer.registerRPSessionTerminationCallback(sessionTermCallback);
                
                // Store viewer globally for debugging
                window.rpViewer = viewer;
                
                // Pre-accept certificate by opening the BMC URL
                handleCertificateAcceptance(viewer);
                
                // Connect after a short delay to allow certificate pre-acceptance
                updateStatus('Accepting BMC certificate and connecting to ' + config.bmcIP + ':' + config.rpPort + '...');
                console.log('Preparing to connect...');
                
                setTimeout(() => {
                    console.log('Calling connectRPViewer...');
                    viewer.connectRPViewer();
                }, 1000);
                
            } catch (error) {
                console.error('Initialization error:', error);
                updateStatus('❌ ERROR: ' + error.message + '<br><br>' +
                           'Check browser console (F12) for more details.', true);
            }
        }

        function exitViewerCallback() {
            console.log('Exit viewer callback');
            updateStatus('Console session ended', true);
        }

        function resolutionCallback(width, height) {
            console.log('Resolution:', width + 'x' + height);
        }

        function sessionTermCallback(reason) {
            console.log('Session terminated:', reason);
            const reasons = {
                0: 'Admin termination',
                1: 'Timeout',
                2: 'WebSocket error',
                3: 'Reboot',
                4: 'Upgrade',
                5: 'Preempted by another user',
                6: 'Unshare',
                7: 'Exclusive mode',
                8: 'Out of memory'
            };
            updateStatus('❌ Session terminated: ' + (reasons[reason] || 'Unknown reason'), true);
        }

        function loginResponseCallback(result, info) {
            console.log('Login response:', result);
            if (result === 0) { // RPViewer.RP_LOGIN_RESULT.LOGIN_SUCCESS
                updateStatus('✓ Connected successfully');
                document.getElementById('certInstructions').style.display = 'none';
                setTimeout(() => {
                    statusDiv.style.display = 'none';
                }, 2000);
            } else {
                const errors = {
                    1: 'Login denied',
                    2: 'Invalid user',
                    3: 'Invalid password',
                    4: 'Session in use',
                    5: 'Session full',
                    6: 'Login timeout',
                    7: 'No share available',
                    11: 'Login failed',
                    101: 'WebSocket exception',
                    102: 'Certificate not verified',
                    103: 'Certificate timeout'
                };
                updateStatus('❌ Login failed: ' + (errors[result] || 'Unknown error'), true);
                
                // Show certificate instructions if it's a certificate error
                if (result === 102 || result === 103) {
                    document.getElementById('certInstructions').style.display = 'block';
                }
            }
        }
        
        // Function to open BMC certificate page
        function acceptCertificate() {
            const certUrl = 'https://' + config.bmcIP + ':' + config.rpPort + '/';
            window.open(certUrl, '_blank');
        }
        
        // Function to retry connection
        function retryConnection() {
            if (window.rpViewer) {
                document.getElementById('certInstructions').style.display = 'none';
                updateStatus('Retrying connection...');
                window.rpViewer.connectRPViewer();
            } else {
                location.reload();
            }
        }

        function uiInitCallback() {
            console.log('UI initialized');
            updateStatus('✓ Console initialized');
        }

        function viewerAPIErrorCallback(code, error) {
            console.error('Viewer API error:', code, error);
            updateStatus('❌ Viewer error: ' + error, true);
        }

        // Handle window resize
        window.addEventListener('resize', function() {
            if (window.rpViewer && window.rpViewer.setRPEmbeddedViewerSize) {
                const canvas = document.getElementById('kvmCanvas');
                window.rpViewer.setRPEmbeddedViewerSize(
                    window.innerWidth,
                    window.innerHeight
                );
            }
        });
        
        // Handle certificate acceptance messages from popup windows
        window.addEventListener('message', function(event) {
            console.log('Received message:', event.data);
            
            // Check for various certificate acceptance message formats
            if (event.data === 'CERT_ACCEPTED' || 
                (event.data && event.data.accepted) ||
                (event.data && event.data.type === 'certificate' && event.data.action === 'accept')) {
                
                console.log('Certificate accepted via popup');
                
                // If RPViewer has a method to handle certificate acceptance
                if (window.rpViewer && window.rpViewer.onCertificateAccepted) {
                    try {
                        window.rpViewer.onCertificateAccepted();
                    } catch(e) {
                        console.log('Could not call onCertificateAccepted:', e);
                    }
                }
            }
        });
        
        // Function that might be called by the certificate popup
        window.rpCertAccepted = function() {
            console.log('Certificate accepted callback triggered');
        };
    </script>
</body>
</html>`
