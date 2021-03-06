﻿using System;
using System.ComponentModel.DataAnnotations;
using System.ComponentModel.DataAnnotations.Schema;
using Fileharbor.Common.Database;

namespace Fileharbor.Services.Entities
{
    public enum LibraryType
    {
        Custom = 0,
        Inbox = 1
    }

    [Table("libraries")]
    public class LibraryEntity
    {
        [Key]
        [ColumnName("id")]
        public Guid Id { get; set; }

        [Required]
        [ColumnName("name")]
        public string Name { get; set; }

        [ColumnName("description")]
        public string Description { get; set; }

        [ColumnName("collection_id")]
        public Guid CollectionId { get; set; }

        [ColumnName("type")]
        public LibraryType Type { get; set; }
    }
}