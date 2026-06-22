package stddetailedline

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"entgo.io/ent/schema/mixin"
	"github.com/alpacahq/alpacadecimal"

	"github.com/openmeterio/openmeter/openmeter/billing/models/creditsapplied"
	"github.com/openmeterio/openmeter/openmeter/billing/models/totals"
	"github.com/openmeterio/openmeter/openmeter/productcatalog"
	"github.com/openmeterio/openmeter/pkg/currencyx"
	"github.com/openmeterio/openmeter/pkg/framework/entutils"
)

type Mixin struct {
	mixin.Schema

	CurrencyPostgresSchemaType string
}

func (m Mixin) base() mixinBase {
	return mixinBase{
		currencyPostgresSchemaType: m.CurrencyPostgresSchemaType,
	}
}

func (m Mixin) Fields() []ent.Field {
	base := m.base()
	fields := base.Fields()

	for _, mixin := range base.Mixin() {
		fields = append(fields, mixin.Fields()...)
	}

	return fields
}

func (m Mixin) Indexes() []ent.Index {
	base := m.base()
	indexes := base.Indexes()

	for _, mixin := range base.Mixin() {
		indexes = append(indexes, mixin.Indexes()...)
	}

	return indexes
}

func (m Mixin) Annotations() []schema.Annotation {
	return m.base().Annotations()
}

type mixinBase struct {
	mixin.Schema

	currencyPostgresSchemaType string
}

func (m mixinBase) currencySchemaType() string {
	if m.currencyPostgresSchemaType != "" {
		return m.currencyPostgresSchemaType
	}

	return currencyx.PostgresCodeSchemaType
}

func (mixinBase) Mixin() []ent.Mixin {
	return []ent.Mixin{
		entutils.AnnotationsMixin{},
		entutils.ResourceMixin{},
		totals.Mixin{},
	}
}

func (m mixinBase) Fields() []ent.Field {
	return []ent.Field{
		field.String("currency").
			GoType(currencyx.Code("")).
			NotEmpty().
			Immutable().
			SchemaType(map[string]string{
				dialect.Postgres: m.currencySchemaType(),
			}),

		// TODO: remove these deprecated detailed-line tax fields after the parent-line
		// inherited tax config rollout has been deployed to production.
		field.JSON("tax_config", productcatalog.TaxConfig{}).
			SchemaType(map[string]string{
				dialect.Postgres: "jsonb",
			}).
			Optional().
			Deprecated("detailed lines inherit tax configuration from their parent standard line"),

		field.String("tax_code_id").
			Optional().
			Nillable().
			SchemaType(map[string]string{
				dialect.Postgres: "char(26)",
			}).
			Deprecated("detailed lines inherit tax configuration from their parent standard line"),
		field.Enum("tax_behavior").
			GoType(productcatalog.TaxBehavior("")).
			Optional().
			Nillable().
			Deprecated("detailed lines inherit tax configuration from their parent standard line"),

		field.Time("service_period_start"),
		field.Time("service_period_end"),

		// Quantity is optional as for usage-based billing we can only persist this value,
		// when the invoice is issued
		field.Other("quantity", alpacadecimal.Decimal{}).
			SchemaType(map[string]string{
				dialect.Postgres: "numeric",
			}),

		// ID of the line in the external invoicing app
		// For example, Stripe invoice line item ID
		field.String("invoicing_app_external_id").
			Optional().
			Nillable(),

		// child_unique_reference_id is uniqe per parent line, can be used for upserting
		// and identifying lines created for the same reason (e.g. tiered price tier)
		// between different invoices.
		field.String("child_unique_reference_id").
			NotEmpty(),

		field.Other("per_unit_amount", alpacadecimal.Decimal{}).
			SchemaType(map[string]string{
				dialect.Postgres: "numeric",
			}),
		field.Enum("category").
			GoType(Category("")).
			Default(string(CategoryRegular)),
		field.Enum("payment_term").
			GoType(productcatalog.PaymentTermType("")).
			Default(string(productcatalog.InAdvancePaymentTerm)),

		field.Int("index").
			Optional().
			Nillable(),

		field.JSON("credits_applied", &creditsapplied.CreditsApplied{}).
			SchemaType(map[string]string{
				dialect.Postgres: "jsonb",
			}).
			Optional(),
	}
}

func (mixinBase) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("tax_code_id"),
	}
}

func (mixinBase) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Checks(map[string]string{
			"child_unique_reference_id_not_empty": `child_unique_reference_id <> ''`,
		}),
	}
}
