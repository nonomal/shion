//! `SeaORM` Entity, @generated by sea-orm-codegen 1.1.0-rc.1

use sea_orm::entity::prelude::*;

#[derive(Clone, Debug, PartialEq, DeriveEntityModel, Eq)]
#[sea_orm(table_name = "domain")]
pub struct Model {
    #[sea_orm(primary_key)]
    pub id: i64,
    #[sea_orm(column_type = "Text")]
    pub name: String,
    #[sea_orm(column_type = "Text")]
    pub color: String,
    #[sea_orm(column_type = "Text")]
    pub pattern: String,
    pub sort: i64,
    pub deleted_at: i64,
    pub created_at: i64,
    pub updated_at: i64,
}

#[derive(Copy, Clone, Debug, EnumIter, DeriveRelation)]
pub enum Relation {
    #[sea_orm(has_many = "super::history::Entity")]
    History,
}

impl Related<super::history::Entity> for Entity {
    fn to() -> RelationDef {
        Relation::History.def()
    }
}

impl ActiveModelBehavior for ActiveModel {}
