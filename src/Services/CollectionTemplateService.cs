﻿using System;
using System.Collections.Generic;
using System.Data;
using System.Data.Common;
using System.Globalization;
using System.IdentityModel.Tokens.Jwt;
using System.IO;
using System.Linq;
using System.Security.Claims;
using System.Threading.Tasks;
using Dapper;
using Fileharbor.Common;
using Fileharbor.Common.Configuration;
using Fileharbor.Common.Database;
using Fileharbor.Common.Utilities;
using Fileharbor.Services.Contracts;
using Fileharbor.Services.Entities;
using Fileharbor.Services.Schema;
using JetBrains.Annotations;
using Microsoft.AspNetCore.Hosting;
using Microsoft.Extensions.Logging;
using Microsoft.Extensions.Options;
using Microsoft.IdentityModel.Tokens;
using Newtonsoft.Json;

namespace Fileharbor.Services
{
    [UsedImplicitly]
    public class CollectionTemplateService : ServiceBase, ICollectionTemplateService
    {
        private readonly ILogger<CollectionTemplateService> _logger;
        private readonly IHostingEnvironment _hostingEnvironment;

        public CollectionTemplateService(ILogger<CollectionTemplateService> logger, IHostingEnvironment hostingEnvironment, IDbConnection database)
            : base(database)
        {
            _logger = logger;
            _hostingEnvironment = hostingEnvironment;
        }

        private async Task<IEnumerable<T>> GetIncludesForTemplate<T>(Template template, IEnumerable<string> includes, string templateFolder)
        {
            var results = new List<T>();

            foreach (var include in includes)
            {
                var path = Path.Combine(templateFolder, include);

                if (!File.Exists(path))
                {
                    _logger.LogCritical("Invalid reference for template {0} and include path {1}", template.Id, path);
                    throw new ArgumentOutOfRangeException(); // TODO: Add exception!
                }

                var content = await File.ReadAllTextAsync(path);
                results.AddRange(JsonConvert.DeserializeObject<T[]>(content));
            }

            return results;
        }

        public async Task<Template> GetTemplateByIdAsync(Guid templateId)
        {
            var templates = await GetAvailableTemplatesAsync();
            // TODO: Throw exception if no tempate found!
            return templates.FirstOrDefault(_ => _.Id == templateId);
        }

        public async Task<IEnumerable<Template>> GetTemplatesByLanguageAsync(int language)
        {
            var templates = await GetAvailableTemplatesAsync();
            // TODO: Throw exception if no tempate found!
            return templates.Where(_ => _.Language.LCID == language);
        }

        // TODO: Implement caching for the templates!
        public async Task<IEnumerable<Template>> GetAvailableTemplatesAsync()
        {
            var contentRoot = _hostingEnvironment.ContentRootPath;
            var templatesDirectory = Path.Combine(contentRoot, Constants.Paths.SchemaDirectory,
                Constants.Paths.TemplatesDirectory);

            _logger.LogDebug("Searching for templates in directory {0}", templatesDirectory);

            var templateFolders = Directory.EnumerateDirectories(templatesDirectory);
            var results = new List<Template>();

            foreach (var templateFolder in templateFolders)
            {
                _logger.LogDebug("Processing template folder {0}", templateFolder);
                var templateFile = Directory.EnumerateFiles(templateFolder).FirstOrDefault(_ =>
                    Path.GetFileName(_).StartsWith("template.", StringComparison.InvariantCultureIgnoreCase));

                if (templateFile == null)
                {
                    _logger.LogDebug("No matching template file found in folder {0}", templateFolder);
                }

                try
                {
                    var content = await File.ReadAllTextAsync(templateFile);
                    var result = JsonConvert.DeserializeObject<Template>(content);

                    var languageIdentifier = Path.GetFileNameWithoutExtension(templateFile).Split('.').LastOrDefault();

                    try
                    {
                        var culture = new CultureInfo(languageIdentifier);
                        result.Language = culture;
                    }
                    catch (Exception e)
                    {
                        _logger.LogWarning(e, "Unable to parse language for filename {0}", templateFile);
                        continue;
                    }

                    result.Columns = await GetIncludesForTemplate<Column>(result, result.ColumnIncludes, templateFolder);
                    result.ContentTypes =
                        await GetIncludesForTemplate<ContentType>(result, result.ContentTypeIncludes, templateFolder);

                    results.Add(result);

                    _logger.LogDebug("Parsed template {0} successfully", templateFile);
                }
                catch (Exception e)
                {
                    _logger.LogWarning(e, "Unable to parse template file {0}", templateFile);
                }
            }

            return results;
        }
    }
}
