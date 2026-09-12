import 'package:flutter/material.dart';
import 'package:flutter_form_builder/flutter_form_builder.dart';
import 'package:flutter_staggered_grid_view/flutter_staggered_grid_view.dart';

class FilterMultiSelect<T> extends StatefulWidget {
  const FilterMultiSelect({
    super.key,
    this.initialSelectedOptions,
    required this.options,
    required this.itemDisplayName,
  });

  final List<T>? initialSelectedOptions;

  final List<T> options;

  final String Function(T) itemDisplayName;

  @override
  FilterMultiSelectState<T> createState() {
    return FilterMultiSelectState();
  }
}

class FilterMultiSelectState<T> extends State<FilterMultiSelect<T>> {
  late List<T> filteredOptions = [...widget.options];
  final formKey = GlobalKey<FormBuilderState>();
  List<T> selectedOptions = [];

  @override
  void initState() {
    super.initState();

    if (widget.initialSelectedOptions != null) {
      widget.initialSelectedOptions!.forEach((element) {
        selectedOptions.add(element);
      });
    }
  }

  @override
  void dispose() {
    super.dispose();
  }

  Widget buildFilterBar() {
    return FormBuilderTextField(
      name: "filter",
      decoration: const InputDecoration(labelText: "Filter"),
      onChanged: (String? value) {
        setState(() {
          filteredOptions = widget.options
              .where((element) => widget
                  .itemDisplayName(element)
                  .toLowerCase()
                  .contains(value!.toLowerCase()))
              .toList();
        });
      },
    );
  }

  Widget buildChoiceChip(T option, int index) {
    return ChoiceChip(
      label: Text(widget.itemDisplayName(option)),
      selected: selectedOptions.contains(option),
      onSelected: (bool selected) {
        setState(() {
          if (selected) {
            selectedOptions.add(option);
          } else {
            selectedOptions.remove(option);
          }
        });
      },
    );
  }

  /// Owns its own scrolling, so it must be given a **bounded** height (the
  /// [Expanded] in [build]). It is deliberately not `shrinkWrap`: a
  /// shrink-wrapped grid sizes itself to its content, so its own scroll extent
  /// is zero -- yet it still claims the vertical drag, which then reaches no
  /// scrollable at all and a long catalog becomes unreachable. Dropping it also
  /// restores lazy chip building.
  Widget buildChoiceChipGrid() {
    return MasonryGridView.count(
      itemCount: filteredOptions.length,
      itemBuilder: (BuildContext context, int index) {
        return buildChoiceChip(filteredOptions[index], index);
      },
      crossAxisCount: 3,
    );
  }

  @override
  Widget build(BuildContext context) {
    return FormBuilder(
        key: formKey,
        child: Padding(
          padding: const EdgeInsets.all(10),
          child: Column(
            children: [
              buildFilterBar(),
              const SizedBox(height: 10),
              // Takes the height the filter bar leaves, which is what gives the
              // grid something to scroll against.
              Expanded(child: buildChoiceChipGrid()),
            ],
          ),
        ));
  }
}
