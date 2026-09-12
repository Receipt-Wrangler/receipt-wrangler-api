import 'package:flutter/material.dart';
import 'package:go_router/go_router.dart';

String getGroupId(BuildContext context) {
  // `extra` is whatever the caller passed, so pattern-match rather than cast:
  // this runs from build methods now (the receipt form's group seed), where a
  // route reached with a non-Map extra would otherwise throw on every frame.
  final extra = GoRouterState.of(context).extra;
  final extraMap = extra is Map ? extra : const {};
  return GoRouterState.of(context).pathParameters["groupId"] ??
      extraMap["groupId"] ??
      "0";
}

String? getGroupByIdWithRouter(GoRouter router) {
  return router.routerDelegate.currentConfiguration.pathParameters["groupId"];
}
