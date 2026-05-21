import 'package:flutter/material.dart';

class AppColors {
  static const bg = Color(0xFF2D0A0F);
  static const surface = Color(0xFF3D1518);
  static const header = Color(0xFF6B0F1A);
  static const feltDark = Color(0xFF1B4D3E);
  static const feltLight = Color(0xFF2D6B4F);
  static const gold = Color(0xFFD4AF37);
  static const goldBright = Color(0xFFFFD700);
  static const goldBorder = Color(0xFFC9A227);
  static const goldMuted = Color(0xFF7A5C1E);
  static const textMuted = Color(0xFFB8A99A);
  static const foldRed = Color(0xFF8B1A1A);
  static const raiseNavy = Color(0xFF1C2B4A);
  static const callGreen = Color(0xFF1B4D2E);
  static const full = Color(0xFF5A3A3A);
}

final appTheme = ThemeData(
  scaffoldBackgroundColor: AppColors.bg,
  useMaterial3: true,
  colorScheme: const ColorScheme.dark(
    primary: AppColors.goldBright,
    secondary: AppColors.gold,
    surface: AppColors.surface,
    background: AppColors.bg,
    onPrimary: AppColors.bg,
    onSecondary: AppColors.bg,
    onSurface: AppColors.goldBright,
    onBackground: AppColors.goldBright,
  ),
  appBarTheme: const AppBarTheme(
    backgroundColor: AppColors.header,
    foregroundColor: AppColors.goldBright,
    elevation: 0,
    centerTitle: true,
    titleTextStyle: TextStyle(
      color: AppColors.goldBright,
      fontSize: 18,
      fontWeight: FontWeight.bold,
    ),
  ),
  cardTheme: CardTheme(
    color: AppColors.surface,
    shape: RoundedRectangleBorder(
      borderRadius: BorderRadius.circular(8),
      side: const BorderSide(color: AppColors.goldBorder),
    ),
    elevation: 0,
  ),
  inputDecorationTheme: InputDecorationTheme(
    filled: true,
    fillColor: AppColors.surface,
    border: OutlineInputBorder(
      borderRadius: BorderRadius.circular(8),
      borderSide: const BorderSide(color: AppColors.goldBorder),
    ),
    enabledBorder: OutlineInputBorder(
      borderRadius: BorderRadius.circular(8),
      borderSide: const BorderSide(color: AppColors.goldBorder),
    ),
    focusedBorder: OutlineInputBorder(
      borderRadius: BorderRadius.circular(8),
      borderSide: const BorderSide(color: AppColors.gold, width: 2),
    ),
    labelStyle: const TextStyle(color: AppColors.textMuted),
    prefixIconColor: AppColors.gold,
  ),
  elevatedButtonTheme: ElevatedButtonThemeData(
    style: ElevatedButton.styleFrom(
      backgroundColor: AppColors.foldRed,
      foregroundColor: AppColors.goldBright,
      padding: const EdgeInsets.symmetric(vertical: 14),
      shape: RoundedRectangleBorder(
        borderRadius: BorderRadius.circular(10),
        side: const BorderSide(color: AppColors.gold, width: 2),
      ),
      textStyle: const TextStyle(
        fontSize: 16,
        fontWeight: FontWeight.bold,
      ),
    ),
  ),
  textButtonTheme: TextButtonThemeData(
    style: TextButton.styleFrom(
      foregroundColor: AppColors.goldBright,
    ),
  ),
  bottomNavigationBarTheme: const BottomNavigationBarThemeData(
    backgroundColor: AppColors.header,
    selectedItemColor: AppColors.goldBright,
    unselectedItemColor: AppColors.textMuted,
    type: BottomNavigationBarType.fixed,
    elevation: 0,
  ),
  dividerTheme: const DividerThemeData(
    color: AppColors.goldBorder,
    thickness: 1,
  ),
);
