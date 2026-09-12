// Demo harness for recording the category/tag tap-target GIF (see CLAUDE.md ->
// "Recording the tap-target demo GIF"). Not shipped in the app.
//
// The left panel rebuilds the PRE-FIX widget tree verbatim -- GestureDetector
// inside the InputDecorator, wrapping only the Wrap, with no hit-test behavior.
// The right panel mounts the real, current MultiSelectField. Everything else is
// identical, so a click that lands in one panel and not the other is the fix.
//
// The demo deliberately opens no picker route: the question is which pixels
// reach the field's onTap at all, and a fullscreen sheet would cover the very
// panels being compared. A tap leaves a marker where it landed -- green when
// the field responded to it, red when the click fell into dead space.
//
// Build: flutter build linux --release --target=tool/tap_target_demo.dart
// The app writes every field's on-screen rect to $TAP_DEMO_RECTS (default
// /tmp/tap_demo_rects.json) after first paint, so the recorder can drive the
// mouse to exact coordinates instead of guessing at the layout.

import 'dart:convert';
import 'dart:io';

import 'package:flutter/material.dart';
import 'package:flutter_form_builder/flutter_form_builder.dart';
import 'package:receipt_wrangler_mobile/shared/widgets/multi-select-field.dart';

void main() => runApp(const _DemoApp());

class _Category {
  const _Category(this.name);

  final String name;
}

const _tags = [_Category('Reimbursable'), _Category('Q3')];

const _beforeColor = Color(0xFFE5484D);
const _afterColor = Color(0xFF30A46C);

/// x that separates the two panels, for attributing a click to one of them.
const _panelSplit = 640.0;

class _DemoApp extends StatelessWidget {
  const _DemoApp();

  @override
  Widget build(BuildContext context) {
    return MaterialApp(
      debugShowCheckedModeBanner: false,
      title: 'MultiSelectField tap target',
      theme: ThemeData(
        fontFamily: 'Raleway',
        // The app's global decoration theme (lib/main.dart) is what gives the
        // field its full-width bordered box, so the demo carries it too.
        inputDecorationTheme: const InputDecorationTheme(
          border: OutlineInputBorder(),
        ),
        chipTheme: ChipThemeData(
          shape: RoundedRectangleBorder(
            borderRadius: BorderRadius.circular(50),
          ),
        ),
        primaryColor: const Color(0xFF27B1FF),
        colorScheme: ColorScheme.fromSeed(seedColor: const Color(0xFF27B1FF)),
      ),
      home: const _DemoScreen(),
    );
  }
}

class _DemoScreen extends StatefulWidget {
  const _DemoScreen();

  @override
  State<_DemoScreen> createState() => _DemoScreenState();
}

class _DemoScreenState extends State<_DemoScreen> {
  final _keys = {
    'beforeCategories': GlobalKey(),
    'beforeTags': GlobalKey(),
    'afterCategories': GlobalKey(),
    'afterTags': GlobalKey(),
  };

  final List<_Click> _clicks = [];
  DateTime? _lastOpenedAt;

  @override
  void initState() {
    super.initState();
    WidgetsBinding.instance.addPostFrameCallback((_) => _dumpRects());
  }

  /// Publishes each field's rect in physical screen pixels, so the recorder can
  /// click exact points inside them.
  void _dumpRects() {
    final dpr = MediaQuery.of(context).devicePixelRatio;

    Map<String, double> rectOf(GlobalKey key) {
      final box = key.currentContext!.findRenderObject() as RenderBox;
      final rect = box.localToGlobal(Offset.zero) & box.size;
      return {
        'left': rect.left * dpr,
        'top': rect.top * dpr,
        'right': rect.right * dpr,
        'bottom': rect.bottom * dpr,
      };
    }

    final path =
        Platform.environment['TAP_DEMO_RECTS'] ?? '/tmp/tap_demo_rects.json';
    File(path).writeAsStringSync(jsonEncode({
      'devicePixelRatio': dpr,
      for (final entry in _keys.entries) entry.key: rectOf(entry.value),
    }));
  }

  void _onPickerOpened() => _lastOpenedAt = DateTime.now();

  int _clicksIn(bool after) =>
      _clicks.where((c) => (c.position.dx > _panelSplit) == after).length;

  int _opensIn(bool after) => _clicks
      .where((c) => (c.position.dx > _panelSplit) == after && c.opened)
      .length;

  /// Marks where a click landed and whether the field responded. The tap
  /// resolves on pointer-up, so the verdict is read one beat later.
  void _recordClick(Offset position) {
    final before = _lastOpenedAt;
    Future.delayed(const Duration(milliseconds: 220), () {
      if (!mounted) return;
      setState(() => _clicks.add(_Click(position, _lastOpenedAt != before)));
    });
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      backgroundColor: const Color(0xFFF1F3F5),
      body: Listener(
        behavior: HitTestBehavior.translucent,
        onPointerUp: (event) => _recordClick(event.position),
        child: Stack(
          children: [
            Padding(
              padding: const EdgeInsets.fromLTRB(28, 20, 28, 20),
              child: Column(
                crossAxisAlignment: CrossAxisAlignment.stretch,
                children: [
                  const _Header(),
                  const SizedBox(height: 16),
                  Row(
                    crossAxisAlignment: CrossAxisAlignment.start,
                    children: [
                      Expanded(
                        child: _Panel(
                          accent: _beforeColor,
                          title: 'BEFORE',
                          subtitle: 'GestureDetector inside the '
                              'InputDecorator,\nwrapping only the Wrap',
                          hits: _opensIn(false),
                          clicks: _clicksIn(false),
                          categories: _LegacyMultiSelectField(
                            fieldKey: _keys['beforeCategories']!,
                            label: 'Categories',
                            itemName: 'Categories',
                            initialValue: const [],
                            onTap: _onPickerOpened,
                          ),
                          tags: _LegacyMultiSelectField(
                            fieldKey: _keys['beforeTags']!,
                            label: 'Tags',
                            itemName: 'Tags',
                            initialValue: _tags,
                            onTap: _onPickerOpened,
                          ),
                        ),
                      ),
                      const SizedBox(width: 28),
                      Expanded(
                        child: _Panel(
                          accent: _afterColor,
                          title: 'AFTER',
                          subtitle: 'Opaque GestureDetector around the\n'
                              'InputDecorator',
                          hits: _opensIn(true),
                          clicks: _clicksIn(true),
                          categories: _CurrentMultiSelectField(
                            fieldKey: _keys['afterCategories']!,
                            name: 'afterCategories',
                            label: 'Categories',
                            itemName: 'Categories',
                            initialValue: const [],
                            onTap: _onPickerOpened,
                          ),
                          tags: _CurrentMultiSelectField(
                            fieldKey: _keys['afterTags']!,
                            name: 'afterTags',
                            label: 'Tags',
                            itemName: 'Tags',
                            initialValue: _tags,
                            onTap: _onPickerOpened,
                          ),
                        ),
                      ),
                    ],
                  ),
                ],
              ),
            ),
            for (final click in _clicks) _ClickMarker(click: click),
          ],
        ),
      ),
    );
  }
}

class _Click {
  const _Click(this.position, this.opened);

  final Offset position;
  final bool opened;
}

class _ClickMarker extends StatelessWidget {
  const _ClickMarker({required this.click});

  final _Click click;

  @override
  Widget build(BuildContext context) {
    final color = click.opened ? _afterColor : _beforeColor;
    return Positioned(
      left: click.position.dx - 13,
      top: click.position.dy - 13,
      child: IgnorePointer(
        child: Container(
          width: 26,
          height: 26,
          decoration: BoxDecoration(
            shape: BoxShape.circle,
            color: color.withValues(alpha: 0.25),
            border: Border.all(color: color, width: 2),
          ),
          child: Icon(
            click.opened ? Icons.check : Icons.close,
            size: 14,
            color: color,
          ),
        ),
      ),
    );
  }
}

class _Header extends StatelessWidget {
  const _Header();

  @override
  Widget build(BuildContext context) {
    return Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        const Text(
          'Category / Tag fields — where you can actually tap',
          style: TextStyle(fontSize: 25, fontWeight: FontWeight.w700),
        ),
        const SizedBox(height: 5),
        Text(
          'Every click below is a real click. Green = the picker opened, '
          'red = the click fell into dead space.',
          style: TextStyle(fontSize: 15, color: Colors.grey.shade700),
        ),
      ],
    );
  }
}

class _Panel extends StatelessWidget {
  const _Panel({
    required this.accent,
    required this.title,
    required this.subtitle,
    required this.hits,
    required this.clicks,
    required this.categories,
    required this.tags,
  });

  final Color accent;
  final String title;
  final String subtitle;
  final int hits;
  final int clicks;
  final Widget categories;
  final Widget tags;

  @override
  Widget build(BuildContext context) {
    return Container(
      decoration: BoxDecoration(
        color: Colors.white,
        borderRadius: BorderRadius.circular(14),
        border: Border.all(color: accent, width: 2),
      ),
      child: Column(
        mainAxisSize: MainAxisSize.min,
        crossAxisAlignment: CrossAxisAlignment.stretch,
        children: [
          Container(
            padding: const EdgeInsets.fromLTRB(18, 10, 18, 12),
            decoration: BoxDecoration(
              color: accent,
              borderRadius: const BorderRadius.vertical(
                top: Radius.circular(12),
              ),
            ),
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                Text(
                  title,
                  style: const TextStyle(
                    color: Colors.white,
                    fontSize: 18,
                    fontWeight: FontWeight.w700,
                    letterSpacing: 1.2,
                  ),
                ),
                const SizedBox(height: 2),
                Text(
                  subtitle,
                  style: const TextStyle(color: Colors.white, fontSize: 13),
                ),
              ],
            ),
          ),
          Padding(
            padding: const EdgeInsets.fromLTRB(18, 18, 18, 16),
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.stretch,
              children: [
                categories,
                const SizedBox(height: 18),
                tags,
                const SizedBox(height: 18),
                _Score(accent: accent, hits: hits, clicks: clicks),
              ],
            ),
          ),
        ],
      ),
    );
  }
}

class _Score extends StatelessWidget {
  const _Score({
    required this.accent,
    required this.hits,
    required this.clicks,
  });

  final Color accent;
  final int hits;
  final int clicks;

  @override
  Widget build(BuildContext context) {
    return Row(
      children: [
        Container(
          padding: const EdgeInsets.symmetric(horizontal: 12, vertical: 6),
          decoration: BoxDecoration(
            color: accent.withValues(alpha: 0.12),
            borderRadius: BorderRadius.circular(20),
            border: Border.all(color: accent.withValues(alpha: 0.5)),
          ),
          child: Text(
            'picker opened on $hits of $clicks clicks',
            style: TextStyle(
              fontSize: 14,
              fontWeight: FontWeight.w600,
              color: accent,
            ),
          ),
        ),
      ],
    );
  }
}

/// The real widget, hosted the way the receipt form hosts it.
class _CurrentMultiSelectField extends StatelessWidget {
  const _CurrentMultiSelectField({
    required this.fieldKey,
    required this.name,
    required this.label,
    required this.itemName,
    required this.initialValue,
    required this.onTap,
  });

  final GlobalKey fieldKey;
  final String name;
  final String label;
  final String itemName;
  final List<_Category> initialValue;
  final VoidCallback onTap;

  @override
  Widget build(BuildContext context) {
    return FormBuilder(
      child: MultiSelectField<_Category>(
        key: fieldKey,
        name: name,
        label: label,
        itemName: itemName,
        itemDisplayName: (category) => category.name,
        initialValue: initialValue,
        onTap: onTap,
      ),
    );
  }
}

/// The pre-fix tree, copied verbatim from multi-select-field.dart as it stood
/// before the tap-target fix, so the panels differ only where the real change
/// did. Its FormBuilderField wrapper is dropped because it contributes no
/// layout -- everything visible here is the original.
class _LegacyMultiSelectField extends StatelessWidget {
  const _LegacyMultiSelectField({
    required this.fieldKey,
    required this.label,
    required this.itemName,
    required this.initialValue,
    required this.onTap,
  });

  final GlobalKey fieldKey;
  final String label;
  final String itemName;
  final List<_Category> initialValue;
  final VoidCallback onTap;

  @override
  Widget build(BuildContext context) {
    Widget buildChip(_Category thing) {
      return ChoiceChip(
        label: Text(thing.name),
        selectedColor: Theme.of(context).primaryColor,
        showCheckmark: false,
        selected: true,
        onSelected: (bool selected) => onTap(),
      );
    }

    List<Widget> buildChipList() {
      if (initialValue.isNotEmpty) {
        final widgets = <Widget>[];
        for (final thing in initialValue) {
          const space = SizedBox(width: 5);
          widgets.add(buildChip(thing));
          widgets.add(space);
        }
        return widgets;
      } else {
        return [Text("No $itemName selected")];
      }
    }

    return InputDecorator(
      key: fieldKey,
      decoration: InputDecoration(labelText: label),
      child: GestureDetector(
          child: Wrap(
            children: buildChipList(),
          ),
          onTap: onTap),
    );
  }
}
