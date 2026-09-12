// GENERATED CODE - DO NOT MODIFY BY HAND

part of 'serializers.dart';

// **************************************************************************
// BuiltValueGenerator
// **************************************************************************

Serializers _$serializers = (Serializers().toBuilder()
      ..add($BaseModel.serializer)
      ..add($PagedRequestCommand.serializer)
      ..add(About.serializer)
      ..add(Activity.serializer)
      ..add(AiType.serializer)
      ..add(ApiKeyFilter.serializer)
      ..add(ApiKeyResult.serializer)
      ..add(ApiKeyScope.serializer)
      ..add(ApiKeyView.serializer)
      ..add(AppData.serializer)
      ..add(AssociatedApiKeys.serializer)
      ..add(AssociatedEntityType.serializer)
      ..add(AssociatedGroup.serializer)
      ..add(BulkStatusUpdateCommand.serializer)
      ..add(BulkUserDeleteCommand.serializer)
      ..add(Category.serializer)
      ..add(CategoryView.serializer)
      ..add(ChartGrouping.serializer)
      ..add(CheckEmailConnectivityCommand.serializer)
      ..add(CheckReceiptProcessingSettingsConnectivityCommand.serializer)
      ..add(Claims.serializer)
      ..add(Comment.serializer)
      ..add(CurrencySeparator.serializer)
      ..add(CurrencySymbolPosition.serializer)
      ..add(CustomField.serializer)
      ..add(CustomFieldOption.serializer)
      ..add(CustomFieldType.serializer)
      ..add(CustomFieldValue.serializer)
      ..add(Dashboard.serializer)
      ..add(DeleteAccountCommand.serializer)
      ..add(EncodedImage.serializer)
      ..add(ExportFormat.serializer)
      ..add(FeatureConfig.serializer)
      ..add(FileData.serializer)
      ..add(FileDataView.serializer)
      ..add(FilterOperation.serializer)
      ..add(GetNewRefreshToken200Response.serializer)
      ..add(GetSystemTaskCommand.serializer)
      ..add(Group.serializer)
      ..add(GroupFilter.serializer)
      ..add(GroupMember.serializer)
      ..add(GroupReceiptSettings.serializer)
      ..add(GroupSettings.serializer)
      ..add(GroupSettingsWhiteListEmail.serializer)
      ..add(GroupStatus.serializer)
      ..add(Icon.serializer)
      ..add(ImportType.serializer)
      ..add(InternalErrorResponse.serializer)
      ..add(Item.serializer)
      ..add(ItemStatus.serializer)
      ..add(LoginCommand.serializer)
      ..add(LogoutCommand.serializer)
      ..add(MagicFillCommand.serializer)
      ..add(Notification.serializer)
      ..add(OcrEngine.serializer)
      ..add(PagedActivityRequestCommand.serializer)
      ..add(PagedApiKeyRequestCommand.serializer)
      ..add(PagedData.serializer)
      ..add(PagedDataDataInner.serializer)
      ..add(PagedGroupRequestCommand.serializer)
      ..add(Permission.serializer)
      ..add(PermissionDescriptor.serializer)
      ..add(PermissionScope.serializer)
      ..add(PieChartData.serializer)
      ..add(PieChartDataCommand.serializer)
      ..add(PieChartDataPoint.serializer)
      ..add(Prompt.serializer)
      ..add(QueueName.serializer)
      ..add(QuickScanDefaultPaidByType.serializer)
      ..add(Receipt.serializer)
      ..add(ReceiptPagedRequestCommand.serializer)
      ..add(ReceiptPagedRequestFilter.serializer)
      ..add(ReceiptProcessingSettings.serializer)
      ..add(ReceiptStatus.serializer)
      ..add(ReportColumn.serializer)
      ..add(ReportColumnAggFuncEnum.serializer)
      ..add(ReportColumnKindEnum.serializer)
      ..add(ReportDetail.serializer)
      ..add(ReportDetailModeEnum.serializer)
      ..add(ReportDocument.serializer)
      ..add(ReportPeriod.serializer)
      ..add(ReportPeriodPresetEnum.serializer)
      ..add(ReportPreviewResponse.serializer)
      ..add(ReportRequestCommand.serializer)
      ..add(ReportRequestCommandFormatsEnum.serializer)
      ..add(ReportTemplate.serializer)
      ..add(ReportTemplateGrant.serializer)
      ..add(ReportTemplateOption.serializer)
      ..add(ResetPasswordCommand.serializer)
      ..add(Role.serializer)
      ..add(SearchResult.serializer)
      ..add(SignUpCommand.serializer)
      ..add(SortDirection.serializer)
      ..add(SubjectLineRegex.serializer)
      ..add(SystemEmail.serializer)
      ..add(SystemSettings.serializer)
      ..add(SystemTask.serializer)
      ..add(SystemTaskPagedRequestFilter.serializer)
      ..add(SystemTaskStatus.serializer)
      ..add(SystemTaskType.serializer)
      ..add(Tag.serializer)
      ..add(TagView.serializer)
      ..add(TaskQueueConfiguration.serializer)
      ..add(TokenPair.serializer)
      ..add(UpdateGroupMemberGrantsCommand.serializer)
      ..add(UpdateGroupReceiptSettingsCommand.serializer)
      ..add(UpdateGroupSettingsCommand.serializer)
      ..add(UpdateProfileCommand.serializer)
      ..add(UpsertApiKeyCommand.serializer)
      ..add(UpsertCategoryCommand.serializer)
      ..add(UpsertCommentCommand.serializer)
      ..add(UpsertCustomFieldCommand.serializer)
      ..add(UpsertCustomFieldOptionCommand.serializer)
      ..add(UpsertCustomFieldValueCommand.serializer)
      ..add(UpsertDashboardCommand.serializer)
      ..add(UpsertGroupCommand.serializer)
      ..add(UpsertGroupMemberCommand.serializer)
      ..add(UpsertItemCommand.serializer)
      ..add(UpsertPromptCommand.serializer)
      ..add(UpsertReceiptCommand.serializer)
      ..add(UpsertReceiptProcessingSettingsCommand.serializer)
      ..add(UpsertRoleCommand.serializer)
      ..add(UpsertSystemEmailCommand.serializer)
      ..add(UpsertSystemSettingsCommand.serializer)
      ..add(UpsertTagCommand.serializer)
      ..add(UpsertTaskQueueConfiguration.serializer)
      ..add(UpsertWidgetCommand.serializer)
      ..add(User.serializer)
      ..add(UserPreferences.serializer)
      ..add(UserShortcut.serializer)
      ..add(UserView.serializer)
      ..add(Widget.serializer)
      ..add(WidgetType.serializer)
      ..addBuilderFactory(
          const FullType(BuiltList, const [const FullType(Category)]),
          () => ListBuilder<Category>())
      ..addBuilderFactory(
          const FullType(BuiltList, const [const FullType(Comment)]),
          () => ListBuilder<Comment>())
      ..addBuilderFactory(
          const FullType(BuiltList, const [const FullType(CustomFieldValue)]),
          () => ListBuilder<CustomFieldValue>())
      ..addBuilderFactory(
          const FullType(BuiltList, const [const FullType(FileData)]),
          () => ListBuilder<FileData>())
      ..addBuilderFactory(
          const FullType(BuiltList, const [const FullType(Item)]),
          () => ListBuilder<Item>())
      ..addBuilderFactory(
          const FullType(BuiltList, const [const FullType(Tag)]),
          () => ListBuilder<Tag>())
      ..addBuilderFactory(
          const FullType(BuiltList, const [const FullType(CustomFieldOption)]),
          () => ListBuilder<CustomFieldOption>())
      ..addBuilderFactory(
          const FullType(BuiltList, const [const FullType(Group)]),
          () => ListBuilder<Group>())
      ..addBuilderFactory(
          const FullType(BuiltList, const [const FullType(UserView)]),
          () => ListBuilder<UserView>())
      ..addBuilderFactory(
          const FullType(BuiltList, const [const FullType(Category)]),
          () => ListBuilder<Category>())
      ..addBuilderFactory(
          const FullType(BuiltList, const [const FullType(Tag)]),
          () => ListBuilder<Tag>())
      ..addBuilderFactory(
          const FullType(BuiltList, const [const FullType(Icon)]),
          () => ListBuilder<Icon>())
      ..addBuilderFactory(
          const FullType(BuiltList, const [const FullType(String)]),
          () => ListBuilder<String>())
      ..addBuilderFactory(
          const FullType(BuiltMap, const [
            const FullType(String),
            const FullType(BuiltList, const [const FullType(String)])
          ]),
          () => MapBuilder<String, BuiltList<String>>())
      ..addBuilderFactory(
          const FullType(BuiltMap, const [
            const FullType(String),
            const FullType(BuiltList, const [const FullType(Category)])
          ]),
          () => MapBuilder<String, BuiltList<Category>>())
      ..addBuilderFactory(
          const FullType(BuiltMap, const [
            const FullType(String),
            const FullType(BuiltList, const [const FullType(Tag)])
          ]),
          () => MapBuilder<String, BuiltList<Tag>>())
      ..addBuilderFactory(
          const FullType(BuiltList, const [const FullType(GroupMember)]),
          () => ListBuilder<GroupMember>())
      ..addBuilderFactory(
          const FullType(BuiltList, const [const FullType(Item)]),
          () => ListBuilder<Item>())
      ..addBuilderFactory(
          const FullType(BuiltList, const [const FullType(Category)]),
          () => ListBuilder<Category>())
      ..addBuilderFactory(
          const FullType(BuiltList, const [const FullType(Tag)]),
          () => ListBuilder<Tag>())
      ..addBuilderFactory(
          const FullType(BuiltList, const [const FullType(PagedDataDataInner)]),
          () => ListBuilder<PagedDataDataInner>())
      ..addBuilderFactory(
          const FullType(BuiltList, const [const FullType(Permission)]),
          () => ListBuilder<Permission>())
      ..addBuilderFactory(
          const FullType(BuiltList, const [const FullType(int)]),
          () => ListBuilder<int>())
      ..addBuilderFactory(
          const FullType(BuiltList, const [const FullType(int)]),
          () => ListBuilder<int>())
      ..addBuilderFactory(
          const FullType(BuiltList, const [const FullType(int)]),
          () => ListBuilder<int>())
      ..addBuilderFactory(
          const FullType(
              BuiltList, const [const FullType(ReportTemplateGrant)]),
          () => ListBuilder<ReportTemplateGrant>())
      ..addBuilderFactory(
          const FullType(BuiltList, const [const FullType(Permission)]),
          () => ListBuilder<Permission>())
      ..addBuilderFactory(
          const FullType(BuiltList, const [const FullType(int)]),
          () => ListBuilder<int>())
      ..addBuilderFactory(
          const FullType(BuiltList, const [const FullType(int)]),
          () => ListBuilder<int>())
      ..addBuilderFactory(
          const FullType(BuiltList, const [const FullType(int)]),
          () => ListBuilder<int>())
      ..addBuilderFactory(
          const FullType(
              BuiltList, const [const FullType(ReportTemplateGrant)]),
          () => ListBuilder<ReportTemplateGrant>())
      ..addBuilderFactory(
          const FullType(BuiltList, const [const FullType(PieChartDataPoint)]),
          () => ListBuilder<PieChartDataPoint>())
      ..addBuilderFactory(
          const FullType(BuiltList, const [const FullType(String)]),
          () => ListBuilder<String>())
      ..addBuilderFactory(
          const FullType(BuiltList, const [const FullType(String)]),
          () => ListBuilder<String>())
      ..addBuilderFactory(
          const FullType(BuiltList, const [const FullType(String)]),
          () => ListBuilder<String>())
      ..addBuilderFactory(
          const FullType(BuiltList, const [const FullType(String)]),
          () => ListBuilder<String>())
      ..addBuilderFactory(
          const FullType(BuiltList, const [const FullType(String)]),
          () => ListBuilder<String>())
      ..addBuilderFactory(
          const FullType(BuiltList, const [const FullType(String)]),
          () => ListBuilder<String>())
      ..addBuilderFactory(
          const FullType(BuiltList, const [const FullType(String)]),
          () => ListBuilder<String>())
      ..addBuilderFactory(
          const FullType(
              BuiltMap, const [const FullType(String), const FullType(String)]),
          () => MapBuilder<String, String>())
      ..addBuilderFactory(
          const FullType(BuiltList, const [const FullType(ReportColumn)]),
          () => ListBuilder<ReportColumn>())
      ..addBuilderFactory(
          const FullType(BuiltList,
              const [const FullType(ReportRequestCommandFormatsEnum)]),
          () => ListBuilder<ReportRequestCommandFormatsEnum>())
      ..addBuilderFactory(
          const FullType(BuiltList, const [const FullType(SubjectLineRegex)]),
          () => ListBuilder<SubjectLineRegex>())
      ..addBuilderFactory(
          const FullType(
              BuiltList, const [const FullType(GroupSettingsWhiteListEmail)]),
          () => ListBuilder<GroupSettingsWhiteListEmail>())
      ..addBuilderFactory(
          const FullType(BuiltList, const [const FullType(SubjectLineRegex)]),
          () => ListBuilder<SubjectLineRegex>())
      ..addBuilderFactory(
          const FullType(
              BuiltList, const [const FullType(GroupSettingsWhiteListEmail)]),
          () => ListBuilder<GroupSettingsWhiteListEmail>())
      ..addBuilderFactory(
          const FullType(BuiltList, const [const FullType(SystemTask)]),
          () => ListBuilder<SystemTask>())
      ..addBuilderFactory(
          const FullType(
              BuiltList, const [const FullType(TaskQueueConfiguration)]),
          () => ListBuilder<TaskQueueConfiguration>())
      ..addBuilderFactory(
          const FullType(
              BuiltList, const [const FullType(UpsertCategoryCommand)]),
          () => ListBuilder<UpsertCategoryCommand>())
      ..addBuilderFactory(
          const FullType(BuiltList, const [const FullType(UpsertTagCommand)]),
          () => ListBuilder<UpsertTagCommand>())
      ..addBuilderFactory(
          const FullType(BuiltList, const [const FullType(UpsertItemCommand)]),
          () => ListBuilder<UpsertItemCommand>())
      ..addBuilderFactory(
          const FullType(
              BuiltList, const [const FullType(UpsertCategoryCommand)]),
          () => ListBuilder<UpsertCategoryCommand>())
      ..addBuilderFactory(
          const FullType(BuiltList, const [const FullType(UpsertTagCommand)]),
          () => ListBuilder<UpsertTagCommand>())
      ..addBuilderFactory(
          const FullType(BuiltList, const [const FullType(UpsertItemCommand)]),
          () => ListBuilder<UpsertItemCommand>())
      ..addBuilderFactory(
          const FullType(
              BuiltList, const [const FullType(UpsertCommentCommand)]),
          () => ListBuilder<UpsertCommentCommand>())
      ..addBuilderFactory(
          const FullType(
              BuiltList, const [const FullType(UpsertCustomFieldValueCommand)]),
          () => ListBuilder<UpsertCustomFieldValueCommand>())
      ..addBuilderFactory(
          const FullType(BuiltList,
              const [const FullType(UpsertCustomFieldOptionCommand)]),
          () => ListBuilder<UpsertCustomFieldOptionCommand>())
      ..addBuilderFactory(
          const FullType(
              BuiltList, const [const FullType(UpsertGroupMemberCommand)]),
          () => ListBuilder<UpsertGroupMemberCommand>())
      ..addBuilderFactory(
          const FullType(
              BuiltList, const [const FullType(UpsertTaskQueueConfiguration)]),
          () => ListBuilder<UpsertTaskQueueConfiguration>())
      ..addBuilderFactory(
          const FullType(
              BuiltList, const [const FullType(UpsertWidgetCommand)]),
          () => ListBuilder<UpsertWidgetCommand>())
      ..addBuilderFactory(
          const FullType(BuiltList, const [const FullType(UserShortcut)]),
          () => ListBuilder<UserShortcut>())
      ..addBuilderFactory(
          const FullType(BuiltList, const [const FullType(Widget)]),
          () => ListBuilder<Widget>())
      ..addBuilderFactory(
          const FullType(BuiltList, const [const FullType(int)]),
          () => ListBuilder<int>())
      ..addBuilderFactory(
          const FullType(BuiltList, const [const FullType(int)]),
          () => ListBuilder<int>())
      ..addBuilderFactory(
          const FullType(BuiltList, const [const FullType(int)]),
          () => ListBuilder<int>())
      ..addBuilderFactory(
          const FullType(BuiltList, const [const FullType(int)]),
          () => ListBuilder<int>())
      ..addBuilderFactory(
          const FullType(BuiltList, const [const FullType(int)]),
          () => ListBuilder<int>())
      ..addBuilderFactory(
          const FullType(BuiltList, const [const FullType(int)]),
          () => ListBuilder<int>())
      ..addBuilderFactory(
          const FullType(BuiltList, const [const FullType(int)]),
          () => ListBuilder<int>())
      ..addBuilderFactory(
          const FullType(BuiltList, const [const FullType(int)]),
          () => ListBuilder<int>())
      ..addBuilderFactory(
          const FullType(BuiltList, const [const FullType(int)]),
          () => ListBuilder<int>())
      ..addBuilderFactory(
          const FullType(BuiltList, const [const FullType(int)]),
          () => ListBuilder<int>())
      ..addBuilderFactory(
          const FullType(BuiltList, const [const FullType(int)]),
          () => ListBuilder<int>())
      ..addBuilderFactory(
          const FullType(BuiltMap, const [
            const FullType(String),
            const FullType.nullable(JsonObject)
          ]),
          () => MapBuilder<String, JsonObject?>())
      ..addBuilderFactory(
          const FullType(BuiltMap, const [
            const FullType(String),
            const FullType.nullable(JsonObject)
          ]),
          () => MapBuilder<String, JsonObject?>()))
    .build();

// ignore_for_file: deprecated_member_use_from_same_package,type=lint
