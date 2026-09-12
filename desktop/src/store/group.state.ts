import { Injectable } from '@angular/core';
import {
  Action,
  createSelector,
  Selector,
  State,
  StateContext,
} from '@ngxs/store';
import { Group } from '../open-api/model/group';
import {
  AddGroup,
  RemoveGroup,
  SetGroups,
  SetSelectedDashboardId,
  SetSelectedGroupId,
  UpdateGroup,
} from './group.state.actions';
import { GroupMember } from '../open-api/model/groupMember';
import { ResetReceiptFilter, SetPage } from './receipt-table.actions';

export interface GroupStateInterface {
  groups: Group[];
  selectedGroupId: string;
  selectedDashboardId: string;
}

@State<GroupStateInterface>({
  name: 'groups',
  defaults: {
    groups: [],
    selectedGroupId: '',
    selectedDashboardId: '',
  },
})
@Injectable()
export class GroupState {
  @Selector()
  static groups(state: GroupStateInterface): Group[] {
    return state.groups;
  }

  @Selector()
  static allGroupMembers(state: GroupStateInterface): GroupMember[] {
    return state.groups.map((g) => g.groupMembers).flat();
  }

  @Selector()
  static groupsWithoutAll(state: GroupStateInterface): Group[] {
    return state.groups.filter((g) => !g.isAllGroup);
  }

  /**
   * The user's only real group, when they belong to exactly one -- a picker with
   * a single option is not a choice, so the receipt form and quick scan seed it
   * rather than making the user pick it. Undefined when they have several (or
   * none): built off `groupsWithoutAll` so the auto-selected set and the set the
   * pickers offer cannot drift apart.
   */
  @Selector([GroupState.groupsWithoutAll])
  static soleGroupId(groups: Group[]): number | undefined {
    return groups.length === 1 ? groups[0].id : undefined;
  }

  /**
   * The group a new receipt would land in when that is not a choice: the
   * actively selected group, or -- when that is the synthetic "All" group, or
   * nothing is selected, or the selection no longer resolves (`selectedGroupId`
   * is persisted, so it outlives a group the user has left) -- the user's only
   * group. Undefined when they genuinely have to pick.
   *
   * Resolved off `groupsWithoutAll`, so "not a real, still-current group" is a
   * single lookup miss rather than three separate checks. Callers that must
   * always name a group (the permission gate, the route guard) fall back to
   * `selectedGroupId` themselves; the form treats undefined as "leave blank".
   */
  @Selector([
    GroupState.groupsWithoutAll,
    GroupState.selectedGroupId,
    GroupState.soleGroupId,
  ])
  static addTargetGroupId(
    selectableGroups: Group[],
    selectedGroupId: string,
    soleGroupId: number | undefined
  ): number | undefined {
    const selected = selectableGroups.find(
      (g) => g.id.toString() === selectedGroupId
    );
    return selected?.id ?? soleGroupId;
  }

  @Selector()
  static groupsWithoutSelectedGroup(state: GroupStateInterface): Group[] {
    return state.groups.filter(
      (g) => g.id.toString() !== state.selectedGroupId
    );
  }

  @Selector()
  static selectedDashboardId(state: GroupStateInterface): string {
    return state.selectedDashboardId;
  }

  @Selector()
  static selectedGroupId(state: GroupStateInterface): string {
    return state.selectedGroupId;
  }

  @Selector()
  static receiptListLink(state: GroupStateInterface): string {
    return `/receipts/group/${state.selectedGroupId}`;
  }

  @Selector()
  static dashboardLink(state: GroupStateInterface): string {
    return `/dashboard/group/${state.selectedGroupId}`;
  }

  @Selector()
  static settingsLinkBase(state: GroupStateInterface): string {
    return `/groups/${state.selectedGroupId}/settings`;
  }

  static getGroupById(groupId: string) {
    return createSelector([GroupState], (state: GroupStateInterface) => {
      return state.groups.find((g) => g.id.toString() === groupId.toString());
    });
  }

  @Action(AddGroup)
  addGroup(
    { getState, patchState }: StateContext<GroupStateInterface>,
    payload: AddGroup
  ) {
    const groups = Array.from(getState().groups);
    groups.push(payload.group);

    patchState({
      groups: groups,
    });
  }

  @Action(RemoveGroup)
  removeGroup(
    { getState, patchState, dispatch }: StateContext<GroupStateInterface>,
    payload: RemoveGroup
  ) {
    const state = getState();
    const group = GroupState.getGroupById(payload.groupId)(state);
    if (group) {
      const index = state.groups.findIndex((g) => g === group);
      if (index >= 0) {
        const newInterface = {} as GroupStateInterface;
        const newGroups = Array.from(state.groups).filter(
          (g) => g.id !== group.id
        );
        newInterface.groups = newGroups;
        const removedSelectedGroup =
          group.id.toString() === state.selectedGroupId.toString();
        if (removedSelectedGroup) {
          // Fall back to the first *remaining* group. Reading state.groups[0]
          // (the pre-removal array) could point at the group just deleted.
          newInterface.selectedGroupId = newGroups[0]?.id?.toString() ?? "";
        }
        patchState(newInterface);

        // Deleting the active group switches the selection to the fallback group,
        // so clear the global receipt-table filter/page the same way a normal
        // group switch does (setSelectedGroupId) — otherwise the fallback group
        // inherits the deleted group's filter.
        if (removedSelectedGroup) {
          return dispatch([new ResetReceiptFilter(), new SetPage(1)]);
        }
      }
    }

    return undefined;
  }

  @Action(SetGroups)
  setGroups(
    { patchState }: StateContext<GroupStateInterface>,
    payload: SetGroups
  ) {
    patchState({
      groups: payload.groups,
    });
  }

  @Action(UpdateGroup)
  updateGroup(
    { getState, patchState }: StateContext<GroupStateInterface>,
    payload: UpdateGroup
  ) {
    const groupIndex = getState().groups.findIndex(
      (g) => g.id?.toString() === payload?.group?.id?.toString()
    );
    if (groupIndex > -1) {
      const newGroups = Array.from(getState().groups);
      newGroups[groupIndex] = payload.group;

      patchState({
        groups: newGroups,
      });
    }
  }

  @Action(SetSelectedDashboardId)
  setSelectedDashboardId(
    { getState, patchState }: StateContext<GroupStateInterface>,
    payload: SetSelectedDashboardId
  ) {
    patchState({
      selectedDashboardId: payload.dashboardId,
    });
  }

  @Action(SetSelectedGroupId)
  setSelectedGroupId(
    { getState, patchState, dispatch }: StateContext<GroupStateInterface>,
    payload: SetSelectedGroupId
  ) {
    const previousGroupId = getState().selectedGroupId;
    let groupId = '';
    let dashboardId = '';

    if (payload?.groupId) {
      groupId = payload.groupId;
    } else {
      const groups = getState().groups;
      groupId = groups[0].id.toString();
    }

    if (groupId === previousGroupId) {
      dashboardId = getState().selectedDashboardId;
    }

    patchState({
      selectedGroupId: groupId,
      selectedDashboardId: dashboardId,
    });

    // The receipt-table state is global (one filter shared across all groups),
    // so a filter set on one group would otherwise bleed into the next. Switching
    // to a different group starts it with a clean filter and first page; columns,
    // page size and sort stay global. Same-group re-selection is a no-op.
    if (groupId !== previousGroupId) {
      return dispatch([new ResetReceiptFilter(), new SetPage(1)]);
    }

    return undefined;
  }
}
