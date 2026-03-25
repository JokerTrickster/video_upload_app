import 'package:flutter/material.dart';

/// Responsive utility based on Galaxy S22+ (412 x 915 dp) as design reference.
/// All sizes scale proportionally to actual device dimensions.
class Responsive {
  static const double _designWidth = 412.0;
  static const double _designHeight = 915.0;

  final BuildContext context;

  Responsive(this.context);

  double get screenWidth => MediaQuery.of(context).size.width;
  double get screenHeight => MediaQuery.of(context).size.height;
  double get statusBarHeight => MediaQuery.of(context).padding.top;
  double get bottomPadding => MediaQuery.of(context).padding.bottom;

  /// Scale width proportionally to design width
  double w(double size) => size * screenWidth / _designWidth;

  /// Scale height proportionally to design height
  double h(double size) => size * screenHeight / _designHeight;

  /// Scale font size (based on width for consistency)
  double sp(double size) => size * screenWidth / _designWidth;

  /// Responsive padding
  EdgeInsets padding({
    double horizontal = 0,
    double vertical = 0,
  }) =>
      EdgeInsets.symmetric(
        horizontal: w(horizontal),
        vertical: h(vertical),
      );

  /// Device type detection
  bool get isSmallPhone => screenWidth < 360;
  bool get isPhone => screenWidth >= 360 && screenWidth < 600;
  bool get isTablet => screenWidth >= 600;

  /// Safe horizontal padding (min 16, scales with screen)
  double get horizontalPadding => w(16).clamp(12.0, 32.0);

  /// Responsive icon size
  double get iconSmall => w(20).clamp(16.0, 24.0);
  double get iconMedium => w(24).clamp(20.0, 32.0);
  double get iconLarge => w(40).clamp(32.0, 56.0);
  double get iconXLarge => w(64).clamp(48.0, 80.0);

  /// Responsive font sizes
  double get bodySmall => sp(12).clamp(10.0, 14.0);
  double get bodyMedium => sp(14).clamp(12.0, 16.0);
  double get bodyLarge => sp(16).clamp(14.0, 18.0);
  double get titleMedium => sp(18).clamp(16.0, 22.0);
  double get titleLarge => sp(22).clamp(18.0, 28.0);
  double get headlineLarge => sp(28).clamp(22.0, 36.0);
}

/// Extension for easy access from BuildContext
extension ResponsiveExtension on BuildContext {
  Responsive get responsive => Responsive(this);
}
