import {
  type AccountPreferencesEnvelope,
  type AccountPreferencesPutRequest,
  type AccountPreferencesResource,
  type AccountProfileEnvelope,
  type AccountProfilePatchRequest,
  type AccountProfileResource,
  type DensityMode,
  type EvidenceAttachBlobEnvelope,
  type EvidenceAttachBlobRequest,
  type EvidenceHandleEnvelope,
  type EvidenceHandleIssueRequest,
  httpOperationBindings,
  type ObjectBlobCreateEnvelope,
  type ObjectBlobCreateRequest,
  type ObjectBlobUploadTarget,
  type ViewCell,
  type ViewMutationData,
  type ViewRow,
} from "@cartulary/protocol-ts";
import type { IncidentStreamMessage } from "@cartulary/protocol-ts/collaboration";
import type {
  AccountPreferencesEnvelope as GeneratedAccountPreferencesEnvelope,
  AccountPreferencesPutRequest as GeneratedAccountPreferencesPutRequest,
  AccountPreferencesResource as GeneratedAccountPreferencesResource,
  AccountProfileEnvelope as GeneratedAccountProfileEnvelope,
  AccountProfilePatchRequest as GeneratedAccountProfilePatchRequest,
  AccountProfileResource as GeneratedAccountProfileResource,
  DensityMode as GeneratedDensityMode,
  EvidenceAttachBlobEnvelope as GeneratedEvidenceAttachBlobEnvelope,
  EvidenceAttachBlobRequest as GeneratedEvidenceAttachBlobRequest,
  EvidenceHandleEnvelope as GeneratedEvidenceHandleEnvelope,
  EvidenceHandleIssueRequest as GeneratedEvidenceHandleIssueRequest,
  ObjectBlobCreateEnvelope as GeneratedObjectBlobCreateEnvelope,
  ObjectBlobCreateRequest as GeneratedObjectBlobCreateRequest,
  ObjectBlobUploadTarget as GeneratedObjectBlobUploadTarget,
  ViewCell as GeneratedViewCell,
  ViewMutationData as GeneratedViewMutationData,
  ViewRow as GeneratedViewRow,
} from "@cartulary/protocol-ts/core-http";
import type { HTTPOperationRequest } from "@cartulary/protocol-ts/http";
import type { TableList } from "@cartulary/protocol-ts/network-flow";

type Assignable<From, To> = [From] extends [To] ? true : false;
type Equal<Left, Right> =
  (<Value>() => Value extends Left ? 1 : 2) extends <
    Value,
  >() => Value extends Right ? 1 : 2
    ? true
    : false;
type Assert<Value extends true> = Value;
type AssertFalse<Value extends false> = Value;

// One bidirectional relationship pair is retained for every handwritten/root
// candidate. These assertions intentionally describe the current compiler
// surface; they are not a substitute for the owner specification.
export type CompatibilityAssignmentMatrix = [
  AssertFalse<Assignable<ViewCell, GeneratedViewCell>>,
  Assert<Assignable<GeneratedViewCell, ViewCell>>,
  AssertFalse<Assignable<ViewRow, GeneratedViewRow>>,
  Assert<Assignable<GeneratedViewRow, ViewRow>>,
  AssertFalse<Assignable<ViewMutationData, GeneratedViewMutationData>>,
  Assert<Assignable<GeneratedViewMutationData, ViewMutationData>>,
  Assert<Assignable<DensityMode, GeneratedDensityMode>>,
  Assert<Assignable<GeneratedDensityMode, DensityMode>>,
  Assert<Assignable<AccountProfileResource, GeneratedAccountProfileResource>>,
  Assert<Assignable<GeneratedAccountProfileResource, AccountProfileResource>>,
  Assert<
    Assignable<AccountPreferencesResource, GeneratedAccountPreferencesResource>
  >,
  Assert<
    Assignable<GeneratedAccountPreferencesResource, AccountPreferencesResource>
  >,
  Assert<
    Assignable<AccountProfilePatchRequest, GeneratedAccountProfilePatchRequest>
  >,
  Assert<
    Assignable<GeneratedAccountProfilePatchRequest, AccountProfilePatchRequest>
  >,
  Assert<
    Assignable<
      AccountPreferencesPutRequest,
      GeneratedAccountPreferencesPutRequest
    >
  >,
  Assert<
    Assignable<
      GeneratedAccountPreferencesPutRequest,
      AccountPreferencesPutRequest
    >
  >,
  Assert<Assignable<AccountProfileEnvelope, GeneratedAccountProfileEnvelope>>,
  Assert<Assignable<GeneratedAccountProfileEnvelope, AccountProfileEnvelope>>,
  Assert<
    Assignable<AccountPreferencesEnvelope, GeneratedAccountPreferencesEnvelope>
  >,
  Assert<
    Assignable<GeneratedAccountPreferencesEnvelope, AccountPreferencesEnvelope>
  >,
  Assert<Assignable<ObjectBlobCreateRequest, GeneratedObjectBlobCreateRequest>>,
  Assert<Assignable<GeneratedObjectBlobCreateRequest, ObjectBlobCreateRequest>>,
  Assert<Assignable<ObjectBlobUploadTarget, GeneratedObjectBlobUploadTarget>>,
  Assert<Assignable<GeneratedObjectBlobUploadTarget, ObjectBlobUploadTarget>>,
  Assert<
    Assignable<ObjectBlobCreateEnvelope, GeneratedObjectBlobCreateEnvelope>
  >,
  Assert<
    Assignable<GeneratedObjectBlobCreateEnvelope, ObjectBlobCreateEnvelope>
  >,
  Assert<
    Assignable<EvidenceAttachBlobRequest, GeneratedEvidenceAttachBlobRequest>
  >,
  Assert<
    Assignable<GeneratedEvidenceAttachBlobRequest, EvidenceAttachBlobRequest>
  >,
  AssertFalse<
    Assignable<EvidenceAttachBlobEnvelope, GeneratedEvidenceAttachBlobEnvelope>
  >,
  Assert<
    Assignable<GeneratedEvidenceAttachBlobEnvelope, EvidenceAttachBlobEnvelope>
  >,
  Assert<
    Assignable<EvidenceHandleIssueRequest, GeneratedEvidenceHandleIssueRequest>
  >,
  Assert<
    Assignable<GeneratedEvidenceHandleIssueRequest, EvidenceHandleIssueRequest>
  >,
  Assert<Assignable<EvidenceHandleEnvelope, GeneratedEvidenceHandleEnvelope>>,
  Assert<Assignable<GeneratedEvidenceHandleEnvelope, EvidenceHandleEnvelope>>,
];

export type DensityModeIsExactlyGenerated = Assert<
  Equal<DensityMode, GeneratedDensityMode>
>;
export type EvidenceHandleIssueRequestIsExactlyGenerated = Assert<
  Equal<EvidenceHandleIssueRequest, GeneratedEvidenceHandleIssueRequest>
>;

export const characterizedCandidates = [
  "ViewCell",
  "ViewRow",
  "ViewMutationData",
  "DensityMode",
  "AccountProfileResource",
  "AccountPreferencesResource",
  "AccountProfilePatchRequest",
  "AccountPreferencesPutRequest",
  "AccountProfileEnvelope",
  "AccountPreferencesEnvelope",
  "ObjectBlobCreateRequest",
  "ObjectBlobUploadTarget",
  "ObjectBlobCreateEnvelope",
  "EvidenceAttachBlobRequest",
  "EvidenceAttachBlobEnvelope",
  "EvidenceHandleIssueRequest",
  "EvidenceHandleEnvelope",
] as const;

export const rootCellWithoutValue: ViewCell = {};
export const rootCellWithExplicitUndefined: ViewCell = { value: undefined };
// @ts-expect-error the generated projection currently requires value
export const generatedCellWithoutValue: GeneratedViewCell = {};
export const generatedCellWithExplicitUndefined: GeneratedViewCell = {
  value: undefined,
};

export const rootCellWithAdditiveMember = {
  future_member: true,
  value: "known",
} satisfies ViewCell;
export const generatedCellWithAdditiveMember = {
  future_member: true,
  value: "known",
} satisfies GeneratedViewCell;

declare const rootCell: ViewCell;
declare const generatedCell: GeneratedViewCell;
export const rootArbitraryLookup: unknown = rootCell["future_member"];
export const generatedArbitraryLookup: unknown = generatedCell["future_member"];

declare const rootProfile: AccountProfileResource;
declare let generatedProfile: GeneratedAccountProfileResource;
// @ts-expect-error the root compatibility declaration is readonly
rootProfile.display_name = "not writable";
generatedProfile.display_name = "generated projection remains writable";

declare const rootUploadTarget: ObjectBlobUploadTarget;
declare const generatedUploadTarget: GeneratedObjectBlobUploadTarget;
export const rootHeaderValue: string | undefined =
  rootUploadTarget.headers["X-Upload"];
export const generatedHeaderValue: string | undefined =
  generatedUploadTarget.headers["X-Upload"];

export const rootClosedHandleRequest = {} satisfies EvidenceHandleIssueRequest;
export const rootHandleRequestWithMember = {
  // @ts-expect-error the root request rejects every member
  unexpected: true,
} satisfies EvidenceHandleIssueRequest;
export const generatedHandleRequestWithMember = {
  // @ts-expect-error the repaired generated request rejects every member
  unexpected: true,
} satisfies GeneratedEvidenceHandleIssueRequest;

function parameterAndReturn<Value>(value: Value): Value {
  return value;
}

export const generatedProfileThroughRootGeneric =
  parameterAndReturn<AccountProfileResource>(generatedProfile);
export const rootProfileThroughGeneratedGeneric =
  parameterAndReturn<GeneratedAccountProfileResource>(rootProfile);
export const rootCellThroughGeneratedGeneric =
  parameterAndReturn<GeneratedViewCell>(
    // @ts-expect-error a root cell may omit value in generic parameter position
    rootCell,
  );
export const generatedCellThroughRootGeneric =
  parameterAndReturn<ViewCell>(generatedCell);

// These declarations make all five supported package specifiers part of the
// real package project compile, including value and type positions.
export const characterizedRootOperation = httpOperationBindings.getIncident;
export type CharacterizedCollaborationSpecifier = IncidentStreamMessage;
export type CharacterizedCoreHTTPSpecifier = GeneratedViewRow;
export type CharacterizedHTTPSpecifier =
  HTTPOperationRequest<"queryWorkbookView">;
export type CharacterizedNetworkFlowSpecifier = TableList;
