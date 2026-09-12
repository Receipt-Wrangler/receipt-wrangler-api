import 'package:flutter/material.dart';
import 'package:flutter_form_builder/flutter_form_builder.dart';
import 'package:form_builder_validators/form_builder_validators.dart';

class MultiSelectField<T> extends StatefulWidget {
  const MultiSelectField(
      {super.key,
      required this.name,
      required this.label,
      required this.itemDisplayName,
      required this.itemName,
      this.initialValue,
      this.onTap,
      this.required});

  final String name;

  final String label;

  final String Function(T) itemDisplayName;

  final String itemName;

  final List<T>? initialValue;

  final Function()? onTap;

  final bool? required;

  @override
  State<MultiSelectField> createState() => _MultiSelectField<T>();
}

class _MultiSelectField<T> extends State<MultiSelectField<T>> {
  @override
  void initState() {
    super.initState();
  }

  @override
  Widget build(BuildContext context) {
    return FormBuilderField<List<T>>(
      name: widget.name,
      validator:
          (widget.required ?? false) ? FormBuilderValidators.required() : null,
      initialValue: widget!.initialValue as dynamic,
      builder: (FormFieldState<List<T>?> field) {
        Widget buildChipLabel(T thing) {
          return Text(widget.itemDisplayName(thing));
        }

        Widget buildChip(T thing) {
          return ChoiceChip(
            label: buildChipLabel(thing),
            selectedColor: Theme.of(context).primaryColor,
            showCheckmark: false,
            selected: true,
            onSelected: (bool selected) {
              if (widget.onTap != null) {
                widget!.onTap!();
              }
            },
          );
        }

        List<Widget> buildChipList() {
          if (field.value != null && field.value!.isNotEmpty) {
            List<Widget> widgets = [];
            for (T thing in field.value!) {
              const space = SizedBox(width: 5);
              widgets.add(buildChip(thing));
              widgets.add(space);
            }
            return widgets;
          } else {
            return [Text("No ${widget.itemName} selected")];
          }
        }

        final decorated = InputDecorator(
          decoration: InputDecoration(labelText: widget.label),
          child: Wrap(
            children: buildChipList(),
          ),
        );

        // View mode has nothing to open, so no tap surface is installed at all
        // -- an opaque detector there would swallow pointers for a no-op.
        if (widget.onTap == null) {
          return decorated;
        }

        // The detector wraps the InputDecorator (not the other way around) and
        // is opaque, so the label, the border and the padding gutters are one
        // tap target. Wrapping only the inner Wrap leaves it shrink-wrapped to
        // its text, and deferToChild drops the gaps between chips.
        return GestureDetector(
          behavior: HitTestBehavior.opaque,
          onTap: widget.onTap,
          child: decorated,
        );
      },
    );
  }
}
