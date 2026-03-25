import 'package:flutter/material.dart';
import 'package:go_router/go_router.dart';
import 'package:provider/provider.dart';
import 'package:url_launcher/url_launcher.dart';
import '../../../core/utils/responsive.dart';
import 'auth_provider.dart';

class LoginScreen extends StatefulWidget {
  const LoginScreen({super.key});

  @override
  State<LoginScreen> createState() => _LoginScreenState();
}

class _LoginScreenState extends State<LoginScreen> {
  @override
  void initState() {
    super.initState();
    WidgetsBinding.instance.addPostFrameCallback((_) {
      _checkAuth();
    });
  }

  Future<void> _checkAuth() async {
    final authProvider = context.read<AuthProvider>();
    await authProvider.checkAuthStatus();
    if (mounted && authProvider.isAuthenticated) {
      context.go('/media');
    }
  }

  Future<void> _signInWithGoogle() async {
    final authProvider = context.read<AuthProvider>();
    try {
      final url = await authProvider.getGoogleAuthUrl();
      final uri = Uri.parse(url);
      if (await canLaunchUrl(uri)) {
        await launchUrl(uri, mode: LaunchMode.externalApplication);
      } else {
        if (mounted) {
          ScaffoldMessenger.of(context).showSnackBar(
            const SnackBar(content: Text('Could not open browser')),
          );
        }
      }
    } catch (e) {
      if (mounted) {
        ScaffoldMessenger.of(context).showSnackBar(
          SnackBar(content: Text('Error: $e')),
        );
      }
    }
  }

  @override
  Widget build(BuildContext context) {
    final r = context.responsive;

    return Scaffold(
      body: SafeArea(
        child: Consumer<AuthProvider>(
          builder: (context, auth, _) {
            return LayoutBuilder(
              builder: (context, constraints) {
                return SingleChildScrollView(
                  child: ConstrainedBox(
                    constraints: BoxConstraints(minHeight: constraints.maxHeight),
                    child: Padding(
                      padding: EdgeInsets.symmetric(horizontal: r.w(32)),
                      child: Column(
                        mainAxisAlignment: MainAxisAlignment.center,
                        children: [
                          SizedBox(height: r.h(100)),
                          Icon(
                            Icons.cloud_upload_rounded,
                            size: r.iconXLarge * 1.5,
                            color: Theme.of(context).colorScheme.primary,
                          ),
                          SizedBox(height: r.h(24)),
                          Text(
                            'Video Upload',
                            style: TextStyle(
                              fontSize: r.headlineLarge,
                              fontWeight: FontWeight.bold,
                            ),
                          ),
                          SizedBox(height: r.h(8)),
                          Text(
                            'YouTube video backup service',
                            style: TextStyle(
                              fontSize: r.bodyLarge,
                              color: Colors.grey[600],
                            ),
                          ),
                          SizedBox(height: r.h(48)),
                          SizedBox(
                            width: double.infinity,
                            height: r.h(52).clamp(44.0, 60.0),
                            child: ElevatedButton.icon(
                              onPressed: auth.isLoading ? null : _signInWithGoogle,
                              icon: auth.isLoading
                                  ? SizedBox(
                                      width: r.iconSmall,
                                      height: r.iconSmall,
                                      child: const CircularProgressIndicator(strokeWidth: 2),
                                    )
                                  : Icon(Icons.login, size: r.iconMedium),
                              label: Text(
                                auth.isLoading ? 'Connecting...' : 'Sign in with Google',
                                style: TextStyle(fontSize: r.bodyLarge),
                              ),
                              style: ElevatedButton.styleFrom(
                                backgroundColor: Theme.of(context).colorScheme.primary,
                                foregroundColor: Colors.white,
                                shape: RoundedRectangleBorder(
                                  borderRadius: BorderRadius.circular(r.w(12)),
                                ),
                              ),
                            ),
                          ),
                          if (auth.error != null) ...[
                            SizedBox(height: r.h(16)),
                            Text(
                              auth.error!,
                              style: TextStyle(color: Colors.red, fontSize: r.bodyMedium),
                              textAlign: TextAlign.center,
                            ),
                          ],
                          SizedBox(height: r.h(100)),
                        ],
                      ),
                    ),
                  ),
                );
              },
            );
          },
        ),
      ),
    );
  }
}
