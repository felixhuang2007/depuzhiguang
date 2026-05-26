import 'package:flutter/material.dart';
import 'package:flutter_bloc/flutter_bloc.dart';
import 'package:flutter_gen/gen_l10n/app_localizations.dart';
import '../blocs/auth_bloc.dart';
import '../theme.dart';
import 'main_screen.dart';

class LoginScreen extends StatefulWidget {
  const LoginScreen({super.key});

  @override
  State<LoginScreen> createState() => _LoginScreenState();
}

class _LoginScreenState extends State<LoginScreen> {
  final _usernameCtrl = TextEditingController();
  final _passwordCtrl = TextEditingController();
  bool _obscurePassword = true;

  @override
  Widget build(BuildContext context) {
    final l10n = AppLocalizations.of(context)!;

    return BlocListener<AuthBloc, AuthState>(
      listener: (context, state) {
        if (state is AuthAuthenticated) {
          Navigator.of(context).pushReplacement(
            MaterialPageRoute(builder: (_) => const MainScreen()),
          );
        } else if (state is AuthError) {
          ScaffoldMessenger.of(context).showSnackBar(
            SnackBar(
              content: Text(state.message),
              backgroundColor: AppColors.foldRed,
            ),
          );
        }
      },
      child: Scaffold(
        body: Container(
          decoration: const BoxDecoration(
            gradient: LinearGradient(
              begin: Alignment.topCenter,
              end: Alignment.bottomCenter,
              colors: [
                AppColors.bg,
                AppColors.surface,
                AppColors.bg,
              ],
            ),
          ),
          child: SafeArea(
            child: Center(
              child: SingleChildScrollView(
                padding: const EdgeInsets.symmetric(horizontal: 32),
                child: BlocBuilder<AuthBloc, AuthState>(
                  builder: (context, state) {
                    final isLoading = state is AuthLoading;

                    return Column(
                      mainAxisAlignment: MainAxisAlignment.center,
                      children: [
                        // Brand icon / logo placeholder
                        GestureDetector(
                          onLongPress: () {
                            _usernameCtrl.text = 'testplayer2';
                            _passwordCtrl.text = 'Test@1234';
                            ScaffoldMessenger.of(context).showSnackBar(
                              const SnackBar(
                                content: Text('已自动填充测试账号'),
                                duration: Duration(seconds: 1),
                              ),
                            );
                          },
                          child: Container(
                            width: 80,
                            height: 80,
                            decoration: BoxDecoration(
                              shape: BoxShape.circle,
                              border: Border.all(
                                color: AppColors.goldBorder,
                                width: 2,
                              ),
                              color: AppColors.header,
                            ),
                            child: const Icon(
                              Icons.casino,
                              size: 40,
                              color: AppColors.goldBright,
                            ),
                          ),
                        ),
                        const SizedBox(height: 24),
                        // Brand title
                        Text(
                          l10n.appTitle,
                          style: const TextStyle(
                            fontSize: 28,
                            fontWeight: FontWeight.bold,
                            color: AppColors.goldBright,
                            letterSpacing: 2,
                          ),
                        ),
                        const SizedBox(height: 8),
                        // Subtitle
                        const Text(
                          'MYANMAR TEXAS HOLD\'EM',
                          style: TextStyle(
                            fontSize: 12,
                            color: AppColors.goldMuted,
                            letterSpacing: 4,
                          ),
                        ),
                        const SizedBox(height: 48),
                        // Username field
                        TextField(
                          controller: _usernameCtrl,
                          enabled: !isLoading,
                          style: const TextStyle(color: AppColors.goldBright),
                          decoration: InputDecoration(
                            labelText: l10n.username,
                            prefixIcon: const Icon(
                              Icons.person_outline,
                              color: AppColors.gold,
                            ),
                          ),
                        ),
                        const SizedBox(height: 20),
                        // Password field
                        TextField(
                          controller: _passwordCtrl,
                          enabled: !isLoading,
                          obscureText: _obscurePassword,
                          style: const TextStyle(color: AppColors.goldBright),
                          decoration: InputDecoration(
                            labelText: l10n.password,
                            prefixIcon: const Icon(
                              Icons.lock_outline,
                              color: AppColors.gold,
                            ),
                            suffixIcon: IconButton(
                              icon: Icon(
                                _obscurePassword
                                    ? Icons.visibility_off
                                    : Icons.visibility,
                                color: AppColors.gold,
                              ),
                              onPressed: () {
                                setState(() {
                                  _obscurePassword = !_obscurePassword;
                                });
                              },
                            ),
                          ),
                        ),
                        const SizedBox(height: 12),
                        // Forgot password
                        Align(
                          alignment: Alignment.centerRight,
                          child: TextButton(
                            onPressed: isLoading ? null : () {},
                            child: Text(
                              l10n.forgotPassword,
                              style: const TextStyle(
                                color: AppColors.textMuted,
                                fontSize: 13,
                              ),
                            ),
                          ),
                        ),
                        const SizedBox(height: 24),
                        // Login button
                        SizedBox(
                          width: double.infinity,
                          height: 48,
                          child: ElevatedButton(
                            onPressed: isLoading
                                ? null
                                : () {
                                    context.read<AuthBloc>().add(
                                          AuthLoginRequested(
                                            _usernameCtrl.text.trim(),
                                            _passwordCtrl.text,
                                          ),
                                        );
                                  },
                            style: ElevatedButton.styleFrom(
                              backgroundColor: AppColors.foldRed,
                              foregroundColor: AppColors.goldBright,
                              disabledBackgroundColor:
                                  AppColors.foldRed.withOpacity(0.5),
                              shape: RoundedRectangleBorder(
                                borderRadius: BorderRadius.circular(10),
                                side: const BorderSide(
                                  color: AppColors.gold,
                                  width: 2,
                                ),
                              ),
                            ),
                            child: isLoading
                                ? const SizedBox(
                                    width: 20,
                                    height: 20,
                                    child: CircularProgressIndicator(
                                      strokeWidth: 2,
                                      color: AppColors.goldBright,
                                    ),
                                  )
                                : Text(
                                    l10n.login,
                                    style: const TextStyle(
                                      fontSize: 16,
                                      fontWeight: FontWeight.bold,
                                    ),
                                  ),
                          ),
                        ),
                        const SizedBox(height: 24),
                        // Register link
                        Row(
                          mainAxisAlignment: MainAxisAlignment.center,
                          children: [
                            Text(
                              l10n.noAccount,
                              style: const TextStyle(
                                color: AppColors.textMuted,
                              ),
                            ),
                            TextButton(
                              onPressed: isLoading ? null : () {},
                              child: Text(
                                l10n.loginNow,
                                style: const TextStyle(
                                  color: AppColors.goldBright,
                                  fontWeight: FontWeight.bold,
                                ),
                              ),
                            ),
                          ],
                        ),
                        const SizedBox(height: 32),
                        // Social login divider
                        Row(
                          children: [
                            const Expanded(
                              child: Divider(color: AppColors.goldBorder),
                            ),
                            Padding(
                              padding:
                                  const EdgeInsets.symmetric(horizontal: 16),
                              child: Text(
                                l10n.otherLogin,
                                style: const TextStyle(
                                  color: AppColors.textMuted,
                                  fontSize: 12,
                                ),
                              ),
                            ),
                            const Expanded(
                              child: Divider(color: AppColors.goldBorder),
                            ),
                          ],
                        ),
                        const SizedBox(height: 20),
                        // Social login buttons
                        Row(
                          mainAxisAlignment: MainAxisAlignment.center,
                          children: [
                            _SocialButton(
                              icon: Icons.chat_bubble_outline,
                              onTap: isLoading ? null : () {},
                            ),
                            const SizedBox(width: 24),
                            _SocialButton(
                              icon: Icons.phone_android,
                              onTap: isLoading ? null : () {},
                            ),
                          ],
                        ),
                        const SizedBox(height: 32),
                      ],
                    );
                  },
                ),
              ),
            ),
          ),
        ),
      ),
    );
  }

  @override
  void dispose() {
    _usernameCtrl.dispose();
    _passwordCtrl.dispose();
    super.dispose();
  }
}

class _SocialButton extends StatelessWidget {
  final IconData icon;
  final VoidCallback? onTap;

  const _SocialButton({required this.icon, this.onTap});

  @override
  Widget build(BuildContext context) {
    return InkWell(
      onTap: onTap,
      borderRadius: BorderRadius.circular(12),
      child: Container(
        width: 48,
        height: 48,
        decoration: BoxDecoration(
          borderRadius: BorderRadius.circular(12),
          border: Border.all(color: AppColors.goldBorder),
          color: AppColors.surface,
        ),
        child: Icon(icon, color: AppColors.gold),
      ),
    );
  }
}
