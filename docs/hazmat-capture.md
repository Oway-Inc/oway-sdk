# Hazardous material declarations

Quote and shipment requests accept optional `hazmat` and `hazmatPaper` fields.
Existing non-hazmat requests need no changes. Your account must be enabled for
hazmat intake before using these fields.

`hazmatPaper.entries` contains one row per hazardous material. Keep ordinary freight
in `orderComponents`; do not repeat one material's declaration across unrelated
items. Use `entryId` to retain a row's identity and `sourceItemReference` to relate
it to your own line item. Neither field creates a component relationship.

Omitted or null facts remain unknown. An explicit `false` is a supplied answer.
Partial declarations are accepted as intake, with missing information requiring
Ops follow-up before confirmation. Never remove incomplete rows to make a request
look complete. `hazmat: false` conflicts with a nonempty declaration or an inherited
hazmat quote and is rejected.

## Quote and booking continuity

Quotes retain the material facts. Booking with that quote inherits its declaration
when `hazmatPaper` is omitted or null. Supplying a paper explicitly replaces the
draft. The quote must belong to your account and satisfy the ordinary booking
checks. Repricing preserves the declaration.

A quote does not certify a shipment. Send certification only after the signer has
explicitly confirmed the current facts:

```json
{
  "certification": {
    "signerName": "Example Signer",
    "statement": "DOT_172_204_A1",
    "confirmed": true
  }
}
```

This illustrates the consent fields, not a complete declaration. A signer name
alone remains supported on initial creation for legacy callers. It does not
recertify changed facts. Identical saves preserve the existing certification;
changed facts require fresh consent. The server records the authenticated actor
and time. An order retry with changed facts returns a conflict; update the existing
order through the supported Ops workflow rather than creating another shipment.

## Material facts

Supply the UN/NA identification, proper shipping name, technical names where
applicable, hazard class/division, subsidiary hazards and assigned packing group.
Packing group is nullable for materials without an assigned group.

Package count and type describe the actual load. `totalQuantity` and
`quantityUnit` stay together; supported units are `L`, `kg`, `lb`, and `gal`.
`innerPackagingNetCapacityL`, `outerPackageGrossKg`, and `flashPointC` use the units
in their names. Do not substitute a catalog limit for an actual measurement.
Limited quantity, hazardous substance, reportable quantity and marine pollutant
claims are nullable booleans. Exception eligibility is assessed from the shipment
facts; a catalog limited-quantity flag alone is insufficient.

Use a shared emergency contact only with `emergencyContactCoversAll: true`.
Otherwise provide and explicitly confirm the contact on each applicable entry.
When a provider requires a registrant or contract identifier, provide it alongside
the contact. Catalog contacts are suggestions requiring confirmation.

`responseInformationDocumentIds` refer to authorized documents on that order and
must cover the specific material. They cannot be reused from another order or
supplied as order-owned documents on a quote. Ops can associate response
information for uncataloged materials.

## Responses and rollout

Quote responses return `hazmat` and `hazmatPaper`. Shipment responses also expose
`hazmatFindings`, `hazmatReadiness`, `hazmatRevision`, and `hazmatCertifiedAt`.
Readiness is `not_declared`, `information_needed`, or `ready_for_ops_review`.
Ready for review is not carrier qualification or dispatch approval.

Editors that support declaration updates should pass the last read revision as
`expectedRevision`. A stale revision returns HTTP 409. Older clients may omit it,
but cannot detect edits made before their stale draft was submitted.

Deploy the compatible API before these SDK fields are used, then upgrade producers
gradually. Omitted fields preserve existing behavior. No migration or bulk
certification of historical shipments is required.
